package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/ledongthuc/pdf"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CVAnalysis struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Filename  string             `bson:"filename" json:"filename"`
	Analysis  string             `bson:"analysis" json:"analysis"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

var collection *mongo.Collection

func main() {

	godotenv.Load()
	connectDB()

	r := gin.Default()

	r.MaxMultipartMemory = 10 << 20 // 10MB

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	r.POST("/analyze", analyzeCVHandler)
	r.GET("/history", getHistoryHandler)
	r.DELETE("/history/:id", deleteHandler)

	fmt.Println("Server running on http://localhost:8081")
	r.Run(":8081")
}

func connectDB() {

	uri := os.Getenv("MONGO_URI")

	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))

	if err != nil {
		panic(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		panic("MongoDB not reachable")
	}

	collection = client.Database("cvanalyzer").Collection("analyses")

	fmt.Println("Connected to MongoDB")
}

func extractTextFromPDF(filePath string) (string, error) {

	f, r, err := pdf.Open(filePath)

	if err != nil {
		return "", err
	}

	defer f.Close()

	var text bytes.Buffer

	for i := 1; i <= r.NumPage(); i++ {

		page := r.Page(i)

		if page.V.IsNull() {
			continue
		}

		content, err := page.GetPlainText(nil)

		if err == nil && content != "" {
			text.WriteString(content)
			text.WriteString("\n")
		}
	}

	result := strings.TrimSpace(text.String())

	if result == "" {
		return "", fmt.Errorf("no text found in PDF")
	}

	return result, nil
}

func analyzeWithGemini(cvText string) (string, error) {

	apiKey := os.Getenv("GEMINI_API_KEY")

	if apiKey == "" {
		return "", fmt.Errorf("Gemini API key missing")
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", apiKey)

	prompt := fmt.Sprintf(`You are a professional CV analyzer.

Provide:

1 Candidate Summary
2 Key Skills
3 Experience Level
4 Suitable Job Roles
5 Strengths
6 Areas to Improve
7 Score out of 10

CV:
%s`, cvText)

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Gemini API error: %s", string(body))
	}

	var geminiResp GeminiResponse

	err = json.Unmarshal(body, &geminiResp)

	if err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

func analyzeCVHandler(c *gin.Context) {

	file, err := c.FormFile("cv")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}

	if !strings.HasSuffix(strings.ToLower(file.Filename), ".pdf") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only PDF allowed"})
		return
	}

	filePath := fmt.Sprintf("./temp_%d.pdf", time.Now().UnixNano())

	err = c.SaveUploadedFile(file, filePath)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file save failed"})
		return
	}

	defer os.Remove(filePath)

	cvText, err := extractTextFromPDF(filePath)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF read failed"})
		return
	}

	analysis, err := analyzeWithGemini(cvText)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := CVAnalysis{
		ID:        primitive.NewObjectID(),
		Filename:  file.Filename,
		Analysis:  analysis,
		CreatedAt: time.Now(),
	}

	_, err = collection.InsertOne(context.Background(), result)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func getHistoryHandler(c *gin.Context) {

	opts := options.Find().SetSort(bson.M{"createdAt": -1})

	cursor, err := collection.Find(context.Background(), bson.M{}, opts)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch failed"})
		return
	}

	defer cursor.Close(context.Background())

	var analyses []CVAnalysis

	cursor.All(context.Background(), &analyses)

	c.JSON(http.StatusOK, analyses)
}

func deleteHandler(c *gin.Context) {

	id, err := primitive.ObjectIDFromHex(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	collection.DeleteOne(context.Background(), bson.M{"_id": id})

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}