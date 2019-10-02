package ci

import (
	"os"
	"strconv"
)

type CiService struct {
	PR     PullRequest
	URL    string
	Event  string
	Branch string
}

type PullRequest struct {
	Reversion string
	Number    int
	Body      string
}

func Drone() (ci CiService, err error) {
	ci.PR.Number = 0
	ci.PR.Reversion = os.Getenv("DRONE_COMMIT_SHA")
	ci.URL = os.Getenv("DRONE_BUILD_LINK")
	ci.Event = os.Getenv("DRONE_BUILD_EVENT")
	ci.Branch = os.Getenv("DRONE_BRANCH")
	pr := os.Getenv("DRONE_PULL_REQUEST")
	if pr == "" {
		return ci, err
	}
	ci.PR.Number, err = strconv.Atoi(pr)
	return ci, err
}
