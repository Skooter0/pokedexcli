package main

import (
	"strings"
	"bufio"
	"os"
	"fmt"
	"net/http"
	"io"
	"encoding/json"
	"time"
	"pokedexcli/internal/pokecache"
	"errors"
	"math/rand"
)

type Config struct {
	Next     	*string
	Previous 	*string
	cache		*pokecache.Cache
}

type cliCommand struct {
	name		string
	description	string
	callback 	func(string) error
}

type LocationResponse struct {
	Count    int `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type Pokemon struct {
	Name			string	`json:"name"`
	Height			int		`json:"height"`
	Weight			int		`json:"weight"`
    Stats   []struct {
        BaseStat 	int `json:"base_stat"`
        Stat     	struct {
            Name 	string `json:"name"`
        } `json:"stat"`
    } `json:"stats"`
    Types   []struct {
        Type struct {
            Name 	string `json:"name"`
        } `json:"type"`
    } `json:"types"`
	BaseExperience	int	`json:"base_experience"`
}

var commands map[string]cliCommand
var config Config
var pokedex = make(map[string]Pokemon)

func initCommands() {
	commands = map[string]cliCommand{
		"exit": {
			name:			"exit",
			description:	"Exit the Pokedex",
			callback:		func(string) error { return commandExit() },
		},
		"help": {
			name:			"help",
			description:	"Displays a help message",
			callback:		func(string) error { return commandHelp() },
		},
		"map": {
			name:        	"map",
			description: 	"Displays the next 20 location areas in the Pokemon world",
			callback:    	func(string) error { return commandMap() },
		},
		"mapb": {
			name:        	"mapb",
			description: 	"Displays the previous 20 location areas in the Pokemon world",
			callback:    	func(string) error { return commandMapBack() },
		},
		"explore": {
			name:			"explore",
			description:	"Lists Pokemon found at a location",
			callback:		commandExploreWrapper,
		},
		"catch": {
			name:			"catch",
			description:	"Catch a Pokemon",
			callback:		commandCatch,
		},
		"inspect": {
			name:			"inspect",
			description:	"Inspect a Pokemon",
			callback:		commandInspect,
		},
		"pokedex": {
			name:			"pokedex",
			description:	"Lists Pokemon you have caught",
			callback:		commandPokedex,
		},
	}
}

func commandMap() error {
	if config.Next == nil {
		fmt.Println("You're already on the last page.")
		return nil
	}
	fetchLocationAreas(*config.Next) // Use the "Next" URL to fetch the next 20 locations
	return nil
}

func commandMapBack() error {
	if config.Previous == nil {
		fmt.Println("You're on the first page.")
		return nil
	}
	fetchLocationAreas(*config.Previous) // Use the "Previous" URL to fetch the previous 20 locations
	return nil
}

func commandExploreWrapper(input string) error {
	parts := strings.Split(input, " ")
	if len(parts) < 2 {
		return errors.New("missing location area name")
	}
	
	locationName := parts[1]
	return commandExplore([]string{locationName}...)
}

func commandExplore(args ...string) error {
    if len(args) < 1 {
        return fmt.Errorf("explore command requires an area name argument")
    }

    area_name := args[0]
    url := "https://pokeapi.co/api/v2/location-area/" + area_name

    // Check cache first
    var body []byte
    var err error
    if cached, found := config.cache.Get(url); found {
        fmt.Println("Cache hit!")
        body = cached
    } else {
        fmt.Println("Cache miss, fetching from API...")
        
        // Make the HTTP GET request
        resp, err := http.Get(url)
        if err != nil {
            return fmt.Errorf("error fetching data: %w", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            return fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
        }

        body, err = io.ReadAll(resp.Body)
        if err != nil {
            return fmt.Errorf("error reading response body: %w", err)
        }

        // Cache the response for future lookups
        config.cache.Add(url, body)
    }

    // Parse the JSON response to extract the Pokémon names
    var explorationData struct {
        PokemonEncounters []struct {
            Pokemon struct {
                Name string `json:"name"`
            } `json:"pokemon"`
        } `json:"pokemon_encounters"`
    }

    err = json.Unmarshal(body, &explorationData)
    if err != nil {
        return fmt.Errorf("error parsing JSON: %w", err)
    }

    // Display the names of the Pokémon
    fmt.Println("Found Pokemon:")
    for _, encounter := range explorationData.PokemonEncounters {
        fmt.Printf(" - %s\n", encounter.Pokemon.Name)
    }

    return nil
}

func commandCatch(input string) error {
	parts := strings.Split(input, " ")
    if len(parts) < 2 {
        return errors.New("you must provide a pokemon name")
    }

	pokemonName := parts[1]  // Get just the pokemon name, not the "catch" command
    fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)
    
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", strings.ToLower(pokemonName))
	
	resp, err := http.Get(url)
    if err != nil {
        return err  // Handle the error from http.Get
    }
    defer resp.Body.Close()
	
	if resp.StatusCode == 404 {
		return fmt.Errorf("pokemon %s not found", pokemonName)
	}
	
    var pokemon Pokemon
    if err := json.NewDecoder(resp.Body).Decode(&pokemon); err != nil {
        return err
    }
	
    catchChance := 100 - pokemon.BaseExperience/2
    randNum := rand.Intn(100)

    // Determine if pokemon is caught
    if randNum <= catchChance {
        // Pokemon is caught! 
        pokedex[pokemon.Name] = pokemon
        fmt.Printf("%s was caught!\n", pokemonName)
		fmt.Printf("You may now inspect it with the inspect command.\n")
    } else {
        fmt.Printf("%s escaped!\n", pokemonName)
    }	
	return nil
}	

func commandInspect(input string) error {
    parts := strings.Split(input, " ")
    if len(parts) < 2 {
        return errors.New("you must provide a pokemon name")
    }
    pokemonName := parts[1]

    // Check if pokemon exists in pokedex
    pokemon, ok := pokedex[pokemonName]
    if !ok {
        fmt.Println("you have not caught that pokemon")
        return nil
    }

    // Display pokemon information
    fmt.Printf("Name: %s\n", pokemon.Name)
    fmt.Printf("Height: %d\n", pokemon.Height)
    fmt.Printf("Weight: %d\n", pokemon.Weight)
    
    fmt.Println("Stats:")
    for _, stat := range pokemon.Stats {
        fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
    }
    
    fmt.Println("Types:")
    for _, t := range pokemon.Types {
        fmt.Printf("  - %s\n", t.Type.Name)
    }

    return nil
}

func commandPokedex(parameter string) error{
	fmt.Printf("Your Pokedex:\n")
	for pokemonName := range pokedex {
		fmt.Printf(" - %s\n", pokemonName)
	}
	return nil
}

func main() {
	config = Config{
		cache: pokecache.NewCache(5 * time.Minute),
	}
	initCommands()
	fetchLocationAreas("https://pokeapi.co/api/v2/location-area/")

	scanner := bufio.NewScanner(os.Stdin)
	rand.Seed(time.Now().UnixNano())


	for {
        fmt.Print("Pokedex > ")
        scanner.Scan()                    // wait for user input
        input := scanner.Text()           // get the input as text
        cleaned := cleanInput(input)      // clean the input
        
		if len(cleaned) == 0 {
			continue
		}
		
		commandName := cleaned[0]
		if command, ok := commands[commandName]; ok {
			err := command.callback(input)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Unknown command")
		}
	}
}

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	for _, cmd := range commands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

func cleanInput(text string) []string {
	trimmed := strings.TrimSpace(text)
	words := strings.Fields(trimmed)
    loweredWords := make([]string, len(words))
    for i, word := range words {
        loweredWords[i] = strings.ToLower(word)
    }
    return loweredWords
}

func fetchLocationAreas(url string) {
    if url == "" {
        url = "https://pokeapi.co/api/v2/location-area/"
    }
	
	var body []byte
	var err error

    if cached, found := config.cache.Get(url); found {
		fmt.Println("Cache hit!")
        body = cached
    } else {
		fmt.Println("Cache miss, fetching from API...")
        // If not in cache, make the API call
        resp, err := http.Get(url)
        if err != nil {
            fmt.Println("Error fetching data:", err)
            return
        }
        defer resp.Body.Close()

        body, err = io.ReadAll(resp.Body)
        if err != nil {
            fmt.Println("Error reading response body:", err)
            return
        }

        // Store in cache
        config.cache.Add(url, body)
    }

    var locationResponse LocationResponse
    err = json.Unmarshal(body, &locationResponse)
    if err != nil {
        fmt.Println("Error parsing JSON:", err)
        return
    }

    // Store the "Next" and "Previous" pagination URLs in the config
    config.Next = locationResponse.Next
    config.Previous = locationResponse.Previous

    // Print the names of the locations
    fmt.Println("Location Areas:")
    for _, location := range locationResponse.Results {
        fmt.Println(location.Name)
    }
}

	