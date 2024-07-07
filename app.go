package main

import (
	"bufio"
	"context"
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


func main() {
    var FARparts strings.Builder

    partsFiles,err := filepath.Glob("src/FARhtml/Part_*.html")

    for i := 1; i < len(partsFiles); i++ { partInfo, err := GetPartInfo(i)
        if err != nil {
            break
        }

        FARparts.WriteString(partInfo)
    }

    err = godotenv.Load()


    if err != nil {
        log.Fatal("Error loading .env file")
    }

    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Ask a Question: ")
    text, _ := reader.ReadString('\n')

    API_KEY:= os.Getenv("api_key")

    client := openai.NewClient(API_KEY)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
                    Content: fmt.Sprintf(`Respond with where in the FAR to find the answer to the question asked. Only respond with information provided from this instruction. If information from a specific part is needed,
                    request a part in it's entirety through the tools \n FAR PARTS: \n%s`, FARparts.String()),
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: text,
				},
			},
            Tools: []openai.Tool {
                {
                    Type: openai.ToolTypeFunction,
                    Function: &openai.FunctionDefinition{
                        Name: "GetPart",
                        Description: "Get the entire FAR Part specified",
                        Parameters: jsonschema.Definition{
                            Type: "object",
                            Properties: map[string]jsonschema.Definition{
                                "partNum": {
                                    Type: jsonschema.Integer,
                                    Description: "A specific part in the FAR",
                                },
                            },
                            Required: []string{"partNum"},
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
        fmt.Println(resp.Choices[0].Message.ToolCalls[0].Function.Arguments)
    } else {
        fmt.Println(resp.Choices[0].Message)
    }
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

