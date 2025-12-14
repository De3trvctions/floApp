package nem12

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	model "floapp/models"
	"floapp/standard-library/utility"

	"github.com/beego/beego/v2/core/logs"
)

type GeneratorOptions struct {
	BatchSize int
}

type GeneratorStats struct {
	Records int64
	Inserts int64
}

func UpsertFromPath(path string, opts GeneratorOptions) (GeneratorStats, error) {
	file, err := os.Open(path)
	if err != nil {
		return GeneratorStats{}, err
	}
	defer file.Close()
	return Upsert(file, opts)
}

func Upsert(in io.Reader, opts GeneratorOptions) (GeneratorStats, error) {
	if opts.BatchSize <= 0 {
		opts.BatchSize = 1000
	}

	db := utility.NewDB()
	tx, err := db.Begin()
	if err != nil {
		return GeneratorStats{}, err
	}
	defer tx.Commit()

	scanner := bufio.NewScanner(in)

	buf := make([]byte, 0, 1024*1024) // init buffer size to 1mb
	scanner.Buffer(buf, 10*1024*1024) // each line of data max 10mb

	var (
		currentNMI      string
		intervalMinutes int
		stats           GeneratorStats
		batch           = make([]model.MeterReading, 0, opts.BatchSize)
	)

	// insertion function
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if err := model.UpsertBatchSQL(batch, tx); err != nil {
			logs.Error("300 line flush failed: %v", err)
			return err
		}
		stats.Inserts += int64(len(batch))
		//clear batch slice data
		batch = batch[:0]
		return nil
	}

	//start consume line
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",") //splite lines into slices
		switch fields[0] {
		case "100":
			continue

		case "200":
			if len(fields) < 9 {
				continue
			}
			currentNMI = strings.TrimSpace(fields[1])
			intervalMinutes, _ = strconv.Atoi(strings.TrimSpace(fields[8]))
			stats.Records++

		case "300":
			if currentNMI == "" || intervalMinutes == 0 {
				logs.Error("300 without valid 200 context")
				continue
			}
			intervalCount := 1440 / intervalMinutes
			startIdx := 2
			endIdx := startIdx + intervalCount // skip first 2value, record + date indicator

			if len(fields) < 7+intervalCount {
				logs.Error("300 record too short")
				continue
			}

			baseDate, err := time.Parse("20060102", strings.TrimSpace(fields[1]))
			if err != nil {
				logs.Error("300 date parse failed: %v", err)
				continue
			}

			if len(fields) < endIdx {
				logs.Error("300 record missing interval data: got %d fields, expected at least %d", len(fields), endIdx)
				continue
			}

			for i := startIdx; i < endIdx; i++ {
				val := strings.TrimSpace(fields[i])
				if val == "" {
					continue
				}

				consumptionValue, err := strconv.ParseFloat(val, 64)
				if err != nil {
					continue
				}

				recordedTime := baseDate.Add(time.Duration((i-startIdx)*intervalMinutes) * time.Minute)

				batch = append(batch, model.MeterReading{
					Nmi:         currentNMI,
					Timestamp:   recordedTime,
					Consumption: consumptionValue,
				})

				// Everytime check if reach batch insertion size, if yes go in flush function
				if len(batch) >= opts.BatchSize {
					if err := flush(); err != nil {
						return stats, err
					}
				}
			}
		case "400":
			continue
		case "500":
			continue
		case "900":
			continue
		default:
			return stats, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return stats, err
	}

	if err := flush(); err != nil {
		return stats, err
	}

	return stats, nil
}
