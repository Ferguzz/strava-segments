package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/Ferguzz/go.strava"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

type configData struct {
	UserId      int64
	SegmentId   int
	AccessToken string
}

func main() {
	config, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal("Can't open configuration file: %s\n", err)
	}

	var userInfo configData
	err = json.Unmarshal(config, &userInfo)
	if err != nil {
		log.Fatal("Can't parse configuration file.\n")
	}

	client := strava.NewClient(userInfo.AccessToken)

	segmentService := strava.NewSegmentsService(client)
	segment, err := segmentService.Get(userInfo.SegmentId).Do()
	if err != nil {
		log.Fatal(fmt.Sprintf("Can't get segment: %s\n", err))
	}

	if segment.PRTime == 0 {
		fmt.Println("You have never ridden this segment.")
		return
	}

	filePath := fmt.Sprintf("data/%d.csv", userInfo.SegmentId)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_RDWR, 0666)
	created := false
	if err != nil {
		file, err = os.Create(filePath)
		if err != nil {
			log.Fatal(fmt.Sprintf("Can't open segment data file: %s\n", err))
		}
		created = true
	}
	defer file.Close()

	var startTime int64 = 0
	if created {
		file.WriteString(fmt.Sprintf("name,%s\n", segment.Name))
		file.WriteString(fmt.Sprintf("pr,%d\n", segment.PRTime))
		file.WriteString("date,elapsed_time\n")
		fmt.Println("New segment data.  This may take a while...")
	} else {
		// Get the date and time of the last effort in the file.
		reader := csv.NewReader(file)
		var lastLine []string
		for {
			record, err := reader.Read()
			if err != nil {
				break
			} else {
				lastLine = record
			}
		}

		t, err := time.Parse("2006-01-02 15:04:05 -0700 MST", lastLine[0])
		if err != nil {
			log.Fatal(fmt.Sprintf("Couldn't parse last time in data file: %s\n", err))
		}
		startTime = t.Unix()
	}

	writer := csv.NewWriter(file)

	athletesService := strava.NewAthletesService(client)
	activitiesService := strava.NewActivitiesService(client)

	for {
		activities, err := athletesService.ListActivities(userInfo.UserId).After(startTime).Do()
		if err != nil {
			log.Fatal(fmt.Sprintf("Error getting recent activities: %s\n", err))
		}
		if len(activities) == 0 {
			break
		}
		for _, activity := range activities {
			detail, err := activitiesService.Get(activity.Id).IncludeAllEfforts().Do()
			if err != nil {
				log.Fatal(fmt.Sprintf("Error getting activity detail: %s\n", err))
			}
			efforts := detail.SegmentEfforts
			for _, effort := range efforts {
				if effort.Segment.Id == int64(userInfo.SegmentId) {
					err := writer.Write([]string{effort.StartDate.String(), strconv.Itoa(effort.ElapsedTime)})
					if err != nil {
						log.Fatal(fmt.Sprintf("Couldn't write to data file: %s", err))
					}
					fmt.Printf("Date: %s, Time: %ds\n", effort.StartDate, effort.ElapsedTime)
					startTime = effort.StartDate.Unix()
				}
			}
		}
	}

	writer.Flush()
}
