package main

import (
	"context"
	"math/rand"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/machinebox/graphql"
)

type Response struct {
	Me []struct {
		Username     string `json:"username"`
		USERID       int    `json:"id"`
		Flair        string `json:"flair"`
		Book_count   int    `json:"books_count"`
		Quoted_books []struct {
			Reading_journals []struct {
				Quote string `json:"entry"`
			} `json:"reading_journals"`
			Book struct {
				Book_id       int    `json:"id"`
				Book_title    string `json:"title"`
				Book_subtitle string `json:"subtitle"`
				Contributions []struct {
					Author struct {
						Name string `json:"name"`
					} `json:"author"`
				} `json:"contributions"`
			} `json:"book"`
		} `json:"user_books"`
	} `json:"me"`
}

const apiURL = "https://api.hardcover.app/v1/graphql"

func findLogLevel() log.Level {

	log.SetReportTimestamp(false)
	switch strings.ToLower(os.Getenv("HCQ_INFO_LEVEL")) {
	case "debug":
		return log.DebugLevel
	case "warn", "warning":
		return log.WarnLevel
	default:
		return log.InfoLevel
	}
}

func queryUserInfo(
	ctx context.Context,
	client graphql.Client,
	authToken string,
) (*Response, error) {

	// Define the GraphQL query
	user_info_request := graphql.NewRequest(`
		query MyQuery {
		me {
	 		username
			id
			flair
			books_count
			user_books(
			where: {reading_journals: {entry: {_is_null: false}, event: {_eq: "quote"}}}
			) {
			reading_journals(where: {event: {_eq: "quote"}}) {
				entry
				event
			}
			book {
				id
				subtitle
				title
				contributions {
				author {
					name
				}
				}
			}
			}
		}
		}
    `)
	user_info_request.Header.Set("Authorization", authToken)

	var resp Response

	if err := client.Run(ctx, user_info_request, &resp); err != nil {
		log.Fatal("req_err", "Error making GraphQL request", err)
	}

	return &resp, nil
}

func main() {
	log.SetLevel(findLogLevel())
	client := graphql.NewClient(apiURL)
	ctx := context.Background()
	authToken := os.Getenv("HARDCOVER_API_TOKEN")

	if authToken == "" {
		log.Fatal("HARDCOVER_API_TOKEN is not set")
	}

	user_info_response, _ := queryUserInfo(ctx, *client, authToken)

	quotedBooks := user_info_response.Me[0].Quoted_books

	quoted_book_ids := make([]int, 0, len(quotedBooks))

	for _, quotedBook := range quotedBooks {
		quoted_book_ids = append(quoted_book_ids, quotedBook.Book.Book_id)
	}

	random_book := quotedBooks[rand.Intn(len(quoted_book_ids))].Book

	if len(user_info_response.Me) > 0 {
		log.Info("", user_info_response.Me[0].Username, "Username")
		log.Info("", user_info_response.Me[0].Flair, "Flair")
		log.Debug("", quoted_book_ids, "quoted_book_ids")
		log.Debug("", "random_book_id", random_book.Book_id)
		log.Info("", random_book.Book_title, "random_book_title")
	} else {
		log.Error("No user data received")
		log.Debug("", user_info_response, "user_info_response")
		os.Exit(1)
	}

	log.Infof("Finding random quote from %s", random_book.Book_title)
}
