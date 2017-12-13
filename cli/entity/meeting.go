package entity

import (
	"fmt"
	"os"
	"time"

	"microservice-agenda/cli/errors"
)

// ---------------------------------------------------
// data structures definition
// ---------------------------------------------------

// Meeting one meeting entity
type Meeting struct {
	Title         string
	Participators []string
	StartTime     time.Time
	EndTime       time.Time
	Sponsor       string
}

// Meetings all the meetings
type meetings struct {
	allMeetings  map[string]Meeting            // key: title, value: address of the Meeting entity that has this title
	onesMeetings map[string]map[string]Meeting // key: user name, value: the meetings the user has participated
}

// AllMeetings only one meetings instance can be accessed
var AllMeetings *meetings

// -----------------------------------------------------
// Meeting structure methods definition
// -----------------------------------------------------

// NewMeeting create a new meeting and add to AllMeetings
func NewMeeting(title, start, end string, parts []string) {
	if !(validateTitle(title)) {
		os.Exit(1)
	}
	if !(validateParticipators(parts)) {
		os.Exit(1)
	}
	startTime, ok1 := getTime(start)
	endTime, ok2 := getTime(end)
	if (!ok1) || (!ok2) {
		os.Exit(1)
	}
	if !(validateTime(startTime, endTime)) {
		os.Exit(1)
	}
	if !(validateNoConflicts(parts, startTime, endTime)) {
		os.Exit(1)
	}

	m := Meeting{
		Title:         title,
		Participators: parts,
		StartTime:     startTime,
		EndTime:       endTime,
		Sponsor:       GetCurrentUser().UserName,
	}

	AllMeetings.allMeetings[title] = m
	for _, part := range parts {
		if AllMeetings.onesMeetings[part] == nil {
			AllMeetings.onesMeetings[part] = make(map[string]Meeting)
		}

		AllMeetings.onesMeetings[part][title] = m
	}
	if AllMeetings.onesMeetings[m.Sponsor] == nil {
		AllMeetings.onesMeetings[m.Sponsor] = make(map[string]Meeting)
	}

	AllMeetings.onesMeetings[m.Sponsor][title] = m
}

// check if title has existed
func validateTitle(title string) bool {
	_, exist := AllMeetings.allMeetings[title]
	if exist {
		errors.ErrorMsg(GetCurrentUser().UserName, "meeting \""+title+"\" has existed. expected another title.")
		return false
	}
	return true
}

// check if all the participators have registered
func validateParticipators(parts []string) bool {
	for _, part := range parts {
		flag := false

		for _, user := range users {
			if part == user.UserName {
				flag = true
			}
		}

		if !flag {
			errors.ErrorMsg(GetCurrentUser().UserName, "meeting participator "+part+" has not registered.")
			return false
		}
	}
	return true
}

// check if start time is less than end time
func validateTime(start, end time.Time) bool {
	if start.After(end) || start.Equal(end) {
		errors.ErrorMsg(GetCurrentUser().UserName, "invalid start time, which should be less than end time")
		return false
	}
	return true
}

// check if there are confilts
func validateNoConflicts(parts []string, start, end time.Time) bool {
	for _, part := range parts {
		for _, ms := range AllMeetings.onesMeetings[part] {
			if !(end.Before(ms.StartTime) || end.Equal(ms.StartTime) ||
				start.After(ms.EndTime) || start.Equal(ms.EndTime)) {
				errors.ErrorMsg(GetCurrentUser().UserName, "participator "+part+" has meeting time conflict.")
				return false
			}
		}
	}
	return true
}

// -----------------------------------------------------
// helpful function
// -----------------------------------------------------

// convert string to time.Time
func getTime(t string) (time.Time, bool) {
	tmpTime, err := time.Parse("2006-01-02", t)
	if err != nil {
		errors.ErrorMsg(GetCurrentUser().UserName, "invalid time format: "+t)
		return time.Time{}, false
	}

	return tmpTime, true
}

// RemoveParticipator remove participators from a meeting
func RemoveParticipator(title, name string) {
	for i, part := range AllMeetings.allMeetings[title].Participators {
		if part == name {
			newP := append(AllMeetings.allMeetings[title].Participators[:i], AllMeetings.allMeetings[title].Participators[i+1:]...)
			tmp := AllMeetings.allMeetings[title]
			AllMeetings.allMeetings[title] = Meeting{
				Title:         tmp.Title,
				Participators: newP,
				StartTime:     tmp.StartTime,
				EndTime:       tmp.EndTime,
				Sponsor:       tmp.Sponsor,
			}
			if len(AllMeetings.allMeetings[title].Participators) == 0 {
				delete(AllMeetings.allMeetings, title)
			}
			break
		}
	}

	delete(AllMeetings.onesMeetings[name], title)

	for _, ms := range AllMeetings.onesMeetings {
		for _, m := range ms {
			if m.Title == title {
				for i, part := range m.Participators {
					if part == name {
						m.Participators = append(m.Participators[:i], m.Participators[i+1:]...)
					}
				}
			}
		}
	}
}

// ------------------------------------------------------
// query meetings methods
// ------------------------------------------------------

// GetMeetings show meetings between time interval [start, end]
func GetMeetings(start, end string) {
	startTime, ok1 := getTime(start)
	endTime, ok2 := getTime(end)
	if (!ok1) || (!ok2) {
		os.Exit(1)
	}
	curUser := GetCurrentUser().UserName
	flag := false

	fmt.Println(curUser + "'s meetings between " + start + " and " + end + ": ")
	ms := AllMeetings.onesMeetings[curUser]
	for _, v := range ms {
		if !(v.StartTime.After(endTime) || v.EndTime.Before(startTime)) {
			fmt.Println()
			fmt.Println("-------------------------------")

			flag = true
			fmt.Println("title: " + v.Title)
			fmt.Printf("participators: %v\n", v.Participators)
			fmt.Println("start time: " + v.StartTime.Format("2006-01-02"))
			fmt.Println("end time: " + v.EndTime.Format("2006-01-02"))
			fmt.Println("sponsor: " + v.Sponsor)

			fmt.Println("-------------------------------")
			fmt.Println()
		}
	}

	if !flag {
		fmt.Println("none.")
	}
}

// -----------------------------------------------------
// initial and save methods
// -----------------------------------------------------

// InitAllMeetings initialize AllMeetings
func InitAllMeetings() {
	ms := loadAllMeetings()

	AllMeetings = new(meetings)
	AllMeetings.allMeetings = make(map[string]Meeting)
	AllMeetings.onesMeetings = make(map[string]map[string]Meeting)
	for _, m := range ms {
		var ps []string
		for _, parts := range m.Participators {
			ps = append(ps, parts)
		}
		AllMeetings.allMeetings[m.Title] = Meeting{
			Title:         m.Title,
			Participators: ps,
			StartTime:     m.StartTime,
			EndTime:       m.EndTime,
			Sponsor:       m.Sponsor,
		}

		for _, person := range m.Participators {
			if AllMeetings.onesMeetings[person] == nil {
				AllMeetings.onesMeetings[person] = make(map[string]Meeting)
			}

			AllMeetings.onesMeetings[person][m.Title] = m
		}

		if AllMeetings.onesMeetings[m.Sponsor] == nil {
			AllMeetings.onesMeetings[m.Sponsor] = make(map[string]Meeting)
		}

		AllMeetings.onesMeetings[m.Sponsor][m.Title] = m
	}
}

// SaveAllMeetings save AllMeetings
func SaveAllMeetings() {
	wirteAllMeetings()
}
