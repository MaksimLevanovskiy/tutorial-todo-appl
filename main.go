package main

import (
	"context"
	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

type Todo struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Completed bool               `json:"completed" bson:"completed"`
	Body      string             `json:"body" bson:"body"`
}

var collection *mongo.Collection

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	mongoURL := os.Getenv("MONGODB_CONNECT_STR")
	clientOptions := options.Client().ApplyURI(mongoURL)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("can't connect to DB: %v", err)
	}
	defer client.Disconnect(context.Background())

	if err = client.Ping(context.Background(), nil); err != nil {
		log.Fatalf("can't ping DB: %v", err)
	}

	collection = client.Database("todos_appl").Collection("todos")
	app := fiber.New()

	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", postTodos)
	app.Patch("/api/todos/:id", patchTodos)
	app.Delete("/api/todos/:id", deleteTodos)
	log.Fatal(app.Listen("0.0.0.0:" + port))

}

func getTodos(ctx fiber.Ctx) error {
	var todos []Todo

	res, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}
	defer res.Close(context.Background())

	for res.Next(context.Background()) {
		var todo Todo
		if err = res.Decode(&todo); err != nil {
			return err
		}
		todos = append(todos, todo)
	}
	return ctx.JSON(todos)
}

func postTodos(ctx fiber.Ctx) error {
	var todo Todo
	if err := ctx.Bind().Body(&todo); err != nil {
		return err
	}
	if todo.Body == "" {
		ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Todo body cannot be empty"})
	}

	insertResult, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError)
	}
	todo.ID = insertResult.InsertedID.(primitive.ObjectID)

	return ctx.Status(fiber.StatusCreated).JSON(todo)
}

func patchTodos(ctx fiber.Ctx) error {
	id := ctx.Params("id")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid todo ID"})
	}

	filter := bson.M{
		"_id": objID,
	}
	update := bson.M{
		"$set": bson.M{
			"completed": true,
		},
	}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"success": true})
}

func deleteTodos(ctx fiber.Ctx) error {
	id := ctx.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "empty todos id"})
	}

	filter := bson.M{"_id": objId}
	_, err = collection.DeleteOne(context.Background(), filter, nil)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}
