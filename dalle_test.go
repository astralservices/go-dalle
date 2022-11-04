package dalle_test

import (
	"os"
	"testing"

	"github.com/astralservices/go-dalle"
	"github.com/joho/godotenv"
)

// load .env

func TestMain(m *testing.M) {
	godotenv.Load("./.env")
	m.Run()
}

func TestGenerate(t *testing.T) {
	apiKey := os.Getenv("DALLE_API_KEY")
	client := dalle.NewClient(apiKey)

	data, err := client.Generate("A horse in an elevator", nil, nil, nil, nil)

	if err != nil {
		t.Error(err)
	}

	if len(data) != 1 {
		t.Error("Expected 1 image, got", len(data))
		t.FailNow()
	}

	if data[0].URL == "" {
		t.Error("Expected URL to be populated")
		t.FailNow()
	}

	t.Log(data[0].URL)
}

func TestEdit(t *testing.T) {
	apiKey := os.Getenv("DALLE_API_KEY")
	client := dalle.NewClient(apiKey)

	file, err := os.Open("./test_data/image_edit_original.png")

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	mask, err := os.Open("./test_data/image_edit_mask.png")

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	data, err := client.Edit("a sunlit indoor lounge area with a pool containing a flamingo", file, mask, nil, nil, nil, nil)

	if err != nil {
		t.Error(err)
	}

	if len(data) != 1 {
		t.Error("Expected 1 image, got", len(data))
		t.FailNow()
	}

	if data[0].URL == "" {
		t.Error("Expected URL to be populated")
		t.FailNow()
	}

	t.Log(data[0].URL)
}

func TestVariation(t *testing.T) {
	apiKey := os.Getenv("DALLE_API_KEY")
	client := dalle.NewClient(apiKey)

	file, err := os.Open("./test_data/image_edit_original.png")

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	data, err := client.Variation(file, nil, nil, nil, nil)

	if err != nil {
		t.Error(err)
	}

	if len(data) != 1 {
		t.Error("Expected 1 image, got", len(data))
		t.FailNow()
	}

	if data[0].URL == "" {
		t.Error("Expected URL to be populated")
		t.FailNow()
	}

	t.Log(data[0].URL)
}
