package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/machinebox/graphql"
)

const apiURL = "https://api.hardcover.app/v1/graphql"

var authToken string

type PrettyQuote struct {
	Quote                string
	BookTitle            string
	BookSubTitle         string
	BookAuthor           string
	HardcoverUser        string
	HardcoverFlair       string
	HardcoverProfileLink string
}

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

func init() {
	log.SetReportTimestamp(false)
	switch strings.ToLower(os.Getenv("HCQ_INFO_LEVEL")) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "warn", "warning":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	authToken = os.Getenv("HARDCOVER_API_TOKEN")

	if authToken == "" {
		log.Fatal("HARDCOVER_API_TOKEN is not set")
	}

}

func printQuote(quote PrettyQuote) {

}

func main() {
	client := graphql.NewClient(apiURL)
	ctx := context.Background()

	user_info_response, _ := queryUserInfo(ctx, *client, authToken)

	quotedBooks := user_info_response.Me[0].Quoted_books

	quoted_book_ids := make([]int, 0, len(quotedBooks))

	for _, quotedBook := range quotedBooks {
		quoted_book_ids = append(quoted_book_ids, quotedBook.Book.Book_id)
	}

	random_book_index := rand.Intn(len(quoted_book_ids))
	random_book := quotedBooks[random_book_index].Book

	if len(user_info_response.Me) > 0 {
		log.Info("", "Username", user_info_response.Me[0].Username)
		log.Info("", "Flair", user_info_response.Me[0].Flair)
		log.Debug("", "quoted_book_ids", quoted_book_ids)
		log.Debug("", "random_book_id", random_book.Book_id)
		log.Info("", "random_book_title", random_book.Book_title)
	} else {
		log.Error("No user data received")
		log.Debug("", "user_info_response", user_info_response)
		os.Exit(1)
	}

	log.Infof("Finding random quote from %s", random_book.Book_title)
	number_of_quotes := len(quotedBooks[random_book_index].Reading_journals)
	quote := strings.TrimSpace(
		quotedBooks[random_book_index].Reading_journals[rand.Intn(number_of_quotes)].Quote,
	)
	log.Info(quote)
	exportedQuote := PrettyQuote{
		Quote:          quote,
		BookTitle:      random_book.Book_title,
		BookSubTitle:   random_book.Book_subtitle,
		BookAuthor:     random_book.Contributions[0].Author.Name,
		HardcoverUser:  user_info_response.Me[0].Username,
		HardcoverFlair: user_info_response.Me[0].Flair,
		HardcoverProfileLink: fmt.Sprintf(
			"https://hardcover.app/@%s",
			user_info_response.Me[0].Username,
		),
	}

	printQuote(exportedQuote)
}
