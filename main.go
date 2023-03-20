package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
)

type GH struct {
	client *github.Client
}

func run(teamName1, teamName2, token string) {
	if teamName1 == "" || teamName2 == "" {
		log.Fatalln("Invalid number of arguments! Two team names need to be provided")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	gh := GH{
		client: client,
	}

	inBoth, err := gh.compare(ctx, teamName1, teamName2)
	if err != nil {
		log.Fatalf("Something went wrong:\n%v\n", err)
	}

	fmt.Printf("%-20s %-20s %s\n", "Login", "Name", "Email")

	for _, m := range inBoth {
		login := *m.Login
		var name string
		var email string
		if m.Name == nil {
			name = "empty"
		} else {
			name = *m.Name
		}
		if m.Email == nil {
			email = "empty"
		} else {
			email = *m.Email
		}
		fmt.Printf("%-20s %-20s %s\n", login, name, email)
	}
}

func (gh GH) lookupTeam(ctx context.Context, team_slug string) ([]*github.User, error) {
	split1 := strings.Split(team_slug, "/")
	org := split1[0]
	team := split1[1]
	members, _, err := gh.client.Teams.ListTeamMembersBySlug(ctx, org, team, &github.TeamListTeamMembersOptions{})
	if err != nil {
		return []*github.User{}, err
	}
	return members, err
}

// Lookup team info
// Get team membership of both teams
// return members missing from second team
func (gh GH) compare(ctx context.Context, slug1, slug2 string) ([]*github.User, error) {
	team1members, err := gh.lookupTeam(ctx, slug1)
	if err != nil {
		return []*github.User{}, err
	}

	team2members, err := gh.lookupTeam(ctx, slug2)
	if err != nil {
		return []*github.User{}, err
	}

	var membersInBothTeams []*github.User
	for _, m1 := range team1members {
		present := false
		for _, m2 := range team2members {
			if *m1.Login == *m2.Login {
				present = true
			}
		}
		if !present {
			membersInBothTeams = append(membersInBothTeams, m1)
		}
	}
	return membersInBothTeams, nil
}

func main() {
	var team1 *string = flag.String("t1", "", "The slug of the base GH team to compare membership: (org/teamName)")
	var team2 *string = flag.String("t2", "", "The slug of the targe GH team to compare membership: (org/TeamName)")
	var token *string = flag.String("token", "", "A GitHub auth token")
	flag.Parse()

	run(*team1, *team2, *token)
}
