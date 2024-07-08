package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/AncientHeroX/contracting/farutils"
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var API_KEY string

func main() {
    var FARparts strings.Builder

    partsFiles,err := filepath.Glob("src/FARhtml/Part_*.html")

    for i := 1; i <= len(partsFiles); i++ { 
        partInfo, err := GetPartInfo(i)
        if err != nil {
            break
        }

        FARparts.WriteString(fmt.Sprintf("%s \n",partInfo))
    }

    err = godotenv.Load()


    if err != nil {
        log.Fatal("Error loading .env file")
    }

    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Ask a Question: ")
    user_question, _ := reader.ReadString('\n')

    API_KEY = os.Getenv("api_key")

    client := openai.NewClient(API_KEY)
	resp, err := client.CreateChatCompletion( context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
                    Content: fmt.Sprintf(`You are an expert in the Federal Acquisition Regulations (FAR) system. Using the provided descriptions of the FAR parts, determine the relevant part and section to answer the question. If the descriptions alone are sufficient, provide the answer directly. If the descriptions are not enough, identify the part and section where detailed information can be found and use a function call to get the necessary details.%s\n`, FARparts.String()),
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: user_question,
				},
			},
            Tools: []openai.Tool {
                {
                    Type: openai.ToolTypeFunction,
                    Function: &openai.FunctionDefinition{
                        Name: "GetPart",
                        Description: "Search a FAR part for specific terms",
                        Parameters: jsonschema.Definition{
                            Type: "object",
                            Properties: map[string]jsonschema.Definition{
                                "partNum": {
                                    Type: jsonschema.Integer,
                                    Description: "A specific part in the FAR",
                                },
                                "searchterms": {
                                    Type: jsonschema.String,
                                    Description: "One string of different variations of a search query, each query MUST be separated by commas. Limit search queries to terms or words. Minimum of 6 queries. If entire part is needed request an empty string (ie: '') Example ('foo, bar, foo bar')",
                                },
                            },
                            Required: []string{"partNum", "searchterms"},
                        },
                    },
                },
            },
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

    if resp.Choices[0].FinishReason == "tool_calls" {
        var args map[string]interface{} 

        err := json.Unmarshal([]byte(resp.Choices[0].Message.ToolCalls[0].Function.Arguments), &args)
        if err != nil {
            log.Fatal("Could not unmarshal json")
        }

        var part int
        var search []string

        for key, value := range args{

            if key == "partNum"{
                val, ok := value.(float64)
                fmt.Print("part: ", value, "\n")
                if ok {
                    part = int(val)
                }
            } else if key == "searchterms"{
                fmt.Print("search: ", value, "\n")
                val, ok := value.(string)

                if ok {
                    search = strings.Split(val, ",")
                }
            }
        }

        searchResult, err := PartSearch(part, search)

        if err != nil {
            log.Fatal("Could not get part")
        }

        if len(searchResult) > 0 {
            Summarize(user_question, searchResult)
        } else {
            fmt.Println(part, search)
            fmt.Println("No info found, searching entire part...")

            searchResult, err = PartSearch(part, []string{""})

            if err != nil {
                log.Fatal("Could not get part")
            }
            Summarize(user_question, searchResult)
        }
    } else {
        fmt.Println(resp.Choices[0].Message.Content)
    }

}
func Summarize(question string, searchResult string) {
    client := openai.NewClient(API_KEY)
	resp, err := client.CreateChatCompletion( context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
                    Content: fmt.Sprintf(`You are an expert assistant specialized in the Federal Acquisition Regulations (FAR) system. I will provide you with a specific FAR part or subpart, and you will respond to my questions solely based on the information from that part or subpart. You are not to use any prior knowledge or external information. Additionally, when you provide an answer, cite the exact location in the provided FAR part or subpart by quoting the relevant sections. \n %s`, searchResult),
				},
				{
					Role:    openai.ChatMessageRoleUser,
                    Content: fmt.Sprintf(question),
				},
			},
        },
	)
    if err != nil {
        log.Fatal("Could not ask gpt")
    }

    fmt.Println(resp.Choices[0].Message.Content)

}
func PartSearch(part int, search []string) (string, error){

    var partInfo strings.Builder

    subParts, err := filepath.Glob(fmt.Sprintf("src/FARhtml/Subpart_%d.*.html", part))


    if err != nil {
        return "", err
    }


    for i := 0; i <= len(subParts); i++ { 
        paths, err :=filepath.Glob(fmt.Sprintf("src/FARhtml/%d.%d*.html", part, i))

        if err != nil {
            log.Fatal("Could not find Path")
        }
        for _, path := range paths {
            subpart, err := farutils.SubPartSearch(path, search)

            if err == nil {
                partInfo.WriteString(fmt.Sprintf("%s\n", subpart))
            }
        }

    }
    return partInfo.String(), nil
}

func GetPartInfo(part int) (string, error) {
    title, err := farutils.GetPartTitle(part)

    if err != nil {
        return "", err
    }

    scope, err := farutils.GetPartScope(part)

    if err != nil {
        return title, nil 
    }

    return fmt.Sprintf("%s: %s", title, scope), nil
}

