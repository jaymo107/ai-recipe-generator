package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.co/jaymo107/ai-recipe-generator/ai"
	"github.com/joho/godotenv"
)

type handlerFunc func(w http.ResponseWriter, r *http.Request)

type renderFunc func(tmpl string, data any, w http.ResponseWriter) error

func htmlRenderer(tmpl string, data any, w http.ResponseWriter) error {
	tmplPath := path.Join("templates", tmpl)
	parsed, err := template.ParseFiles(tmplPath)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	if err := parsed.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

func homePageHandler(logger *log.Logger, renderer renderFunc) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			renderer("index.html", nil, w)
			logger.Println("Home page is visited.")
		}
	}
}

func generateHandler(logger *log.Logger, recipeGenerator *ai.RecipeGenerator, renderer renderFunc) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		ingredients := strings.Split(r.FormValue("ingredients"), "\n")

		logger.Println("Generating")

		if len(ingredients) == 0 {
			http.Error(w, "Please provide some ingredients", http.StatusBadRequest)
			return
		}

		recipe, err := recipeGenerator.Generate(ingredients)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		renderer("recipe.html", recipe, w)
		logger.Println("Generate page is visited with data", len(ingredients))
	}
}

func readEnvFile() {
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}
}

func main() {
	readEnvFile()

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Println("Server is starting...")

	recipeGenerator := ai.NewRecipeGenerator(os.Getenv("OPENAI_API_KEY"), logger)

	http.HandleFunc("/generate", generateHandler(logger, recipeGenerator, htmlRenderer))
	http.HandleFunc("/", homePageHandler(logger, htmlRenderer))

	http.ListenAndServe(":8080", nil)
}
