package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/machinebox/graphql"
)

const apiURL = "https://api.hardcover.app/v1/graphql"

var authToken string

type PrettyQuote struct {
	Quote                string `json:"quote"`
	BookTitle            string `json:"book_title"`
	BookSubTitle         string `json:"book_subtitle"`
	BookAuthor           string `json:"book_author"`
	HardcoverUser        string `json:"user"`
	HardcoverFlair       string `json:"flair"`
	HardcoverProfileLink string `json:"profile_link"`
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

func (q PrettyQuote) sendWebhook(url string) error {
	jsonData, err := json.Marshal(struct {
		MergeVariables PrettyQuote `json:"merge_variables"`
	}{MergeVariables: q})
	if err != nil {
		log.Error("json_marshal_error", "Error marshalling PrettyQuote to JSON", err)
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error("http_post_error", "Error sending webhook", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook responded with status: %s", resp.Status)
	}

	return nil
}

func queryUserInfo(
	ctx context.Context,
	client graphql.Client,
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
	user_info_request.Header.Set(
		"User-Agent",
		"hcq - https://github.com/GianniBYoung/hardcover-quotes",
	)

	var resp Response

	if err := client.Run(ctx, user_info_request, &resp); err != nil {
		return nil, fmt.Errorf("error making GraphQL request: %w", err)
	}

	return &resp, nil
}

func init() {
	log.SetReportTimestamp(false)
	switch strings.ToLower(os.Getenv("HCQ_INFO_LEVEL")) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	default:
		log.SetLevel(log.WarnLevel)
	}
}

func printQuote(quote PrettyQuote) {
	var b strings.Builder

	b.WriteString(quote.BookTitle + " - " + quote.BookSubTitle)

	b.WriteString("\n")
	b.WriteString("\tby " + quote.BookAuthor)
	b.WriteString("\n\n~~~~~~~~~~~~~~~\n")
	b.WriteString(quote.Quote)
	b.WriteString("\n~~~~~~~~~~~~~~~\n\n")
	b.WriteString(
		quote.HardcoverUser + " - " + quote.HardcoverFlair + " - " + quote.HardcoverProfileLink,
	)
	fmt.Println(b.String())

}

func main() {
	authToken = os.Getenv("HARDCOVER_API_TOKEN")
	if authToken == "" {
		log.Fatal("HARDCOVER_API_TOKEN is not set")
	}

	authToken = strings.TrimSpace(authToken)
	if !strings.HasPrefix(authToken, "Bearer ") {
		authToken = "Bearer " + authToken
	}

	client := graphql.NewClient(apiURL)
	ctx := context.Background()

	user_info_response, err := queryUserInfo(ctx, *client)
	if err != nil {
		log.Fatal("query error", "err", err)
	}

	if len(user_info_response.Me) == 0 {
		log.Error("No user data received from Hardcover")
		os.Exit(1)
	}

	user := user_info_response.Me[0]
	quotedBooks := user.Quoted_books

	if len(quotedBooks) == 0 {
		log.Warn("No books with quotes found in reading journals")
		os.Exit(0)
	}

	random_book_index := rand.Intn(len(quotedBooks))
	random_book := quotedBooks[random_book_index].Book

	log.Info("", "Username", user.Username)
	log.Info("", "Flair", user.Flair)
	log.Debug("", "random_book_id", random_book.Book_id)
	log.Info("", "random_book_title", random_book.Book_title)

	journals := quotedBooks[random_book_index].Reading_journals
	if len(journals) == 0 {
		log.Warn("Selected book has no quote entries")
		os.Exit(0)
	}

	log.Infof("Finding random quote from %s", random_book.Book_title)
	quote := strings.TrimSpace(journals[rand.Intn(len(journals))].Quote)

	author := "Unknown Author"
	if len(random_book.Contributions) > 0 && random_book.Contributions[0].Author.Name != "" {
		author = random_book.Contributions[0].Author.Name
	}

	exportedQuote := PrettyQuote{
		Quote:                quote,
		BookTitle:            random_book.Book_title,
		BookSubTitle:         random_book.Book_subtitle,
		BookAuthor:           author,
		HardcoverUser:        user.Username,
		HardcoverFlair:       user.Flair,
		HardcoverProfileLink: fmt.Sprintf("https://hardcover.app/@%s", user.Username),
	}

	printQuote(exportedQuote)

	webhookURL := os.Getenv("HCQ_WEBHOOK_URL")
	if webhookURL != "" {
		if err := exportedQuote.sendWebhook(webhookURL); err != nil {
			log.Error("webhook_error", "Error sending webhook", err)
		}
	}
}
