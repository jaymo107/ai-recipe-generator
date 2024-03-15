package ai

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type RecipeGenerator struct {
	client *openai.Client
	logger *log.Logger
}

type Recipe struct {
	Name         string   `json:"name"`
	Ingredients  []string `json:"ingredients"`
	Instructions []string `json:"instructions"`
}

func NewRecipeGenerator(apiKey string, logger *log.Logger) *RecipeGenerator {
	client := openai.NewClient(apiKey)
	return &RecipeGenerator{client: client, logger: logger}
}

func (rg *RecipeGenerator) Generate(ingredients []string) (Recipe, error) {
	formattedIngredients := strings.Join(ingredients, ", ")

	recipe, err := rg.makeRequest(`
		You are a pro home chef, you make simple and delicious recipes that don't require a lot of ingredients or a lot of time.
		I will provide you with a list of ingredients, you don't have to use all of them and you may also use some common houshold staples to make the recipe
		such as salt, pepper, oil, and some herbs and spices.

		All of the recipes must be for 2 people and be vegan, show the quantities of ingredients in the list.

		You have the following ingredients to make a recipe out of for a beginner cook:

		` + formattedIngredients + `

		Provide the recipe in JSON format with the keys of name, ingredients and instructions.
		The keys ingredients and instructions will be arrays of strings.
		Name will be the name of the recipe.
	`)

	if err != nil {
		return Recipe{}, err
	}

	return recipe, nil
}

func (rg *RecipeGenerator) makeRequest(prompt string) (Recipe, error) {
	rg.logger.Println("Making request to OpenAI with prompt:", prompt)

	resp, err := rg.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		rg.logger.Println("Error from OpenAI:", err)
		return Recipe{}, err
	}

	rg.logger.Println("Response from OpenAI:", resp)
	rg.logger.Println(resp.Choices)

	return rg.parseResponse(resp.Choices[0].Message.Content)
}

func (rg *RecipeGenerator) parseResponse(response string) (Recipe, error) {
	rg.logger.Println("Parsing response from OpenAI")

	recipe := Recipe{}
	err := json.Unmarshal([]byte(response), &recipe)

	if err != nil {
		rg.logger.Println("Error parsing response from OpenAI:", err)
		return Recipe{}, err
	}

	rg.logger.Println("Parsed response from OpenAI:", recipe.Name)

	return recipe, nil
}
