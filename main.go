package main

import (
    "os"
    "github.com/joho/godotenv"
    "context"
    "github.com/gofiber/fiber/v2"
    "github.com/jackc/pgx/v5/pgxpool"
    "log"
    "net/http"
)

var db *pgxpool.Pool

type Product struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Price       float64 `json:"price"`
}

// Handler est la fonction exportée attendue par Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
    app := fiber.New()

    // Charger les variables d'environnement
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    // Charger la chaîne de connexion à partir des variables d'environnement
    dbURL := os.Getenv("DATABASE_URL")
    db, err = pgxpool.New(context.Background(), dbURL)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Routes
    app.Get("/products", getProducts)
    app.Get("/products/:id", getProduct)
    app.Post("/products", createProduct)
    app.Put("/products/:id", updateProduct)
    app.Delete("/products/:id", deleteProduct)

    // Vercel attend une réponse HTTP standard
    app.Listener = http.NewServeMux()
    app.Listen(":3000")
}

// Récupérer tous les produits  
func getProducts(c *fiber.Ctx) error {  
    rows, err := db.Query(context.Background(), "SELECT * FROM products")  
    if err != nil {  
        return c.Status(fiber.StatusInternalServerError).SendString(err.Error())  
    }  
    defer rows.Close()  

    products := []Product{}  
    for rows.Next() {  
        var prod Product  
        if err := rows.Scan(&prod.ID, &prod.Name, &prod.Description, &prod.Price); err != nil {  
            return err  
        }  
        products = append(products, prod)  
    }  
    return c.JSON(products)  
}  

// Récupérer un produit par ID  
func getProduct(c *fiber.Ctx) error {  
    id := c.Params("id")  
    var prod Product  
    err := db.QueryRow(context.Background(), "SELECT * FROM products WHERE id=$1", id).Scan(&prod.ID, &prod.Name, &prod.Description, &prod.Price)  
    if err != nil {  
        return c.Status(http.StatusNotFound).SendString(err.Error())  
    }  
    return c.JSON(prod)  
}  

// Créer un nouveau produit  
func createProduct(c *fiber.Ctx) error {  
    var prod Product  
    if err := c.BodyParser(&prod); err != nil {  
        return c.Status(http.StatusBadRequest).SendString(err.Error())  
    }  

    // Insertion dans la base de données  
    _, err := db.Exec(context.Background(), "INSERT INTO products (name, description, price) VALUES ($1, $2, $3)", prod.Name, prod.Description, prod.Price)  
    if err != nil {  
        return c.Status(http.StatusInternalServerError).SendString(err.Error())  
    }  
    // Répondre avec le produit créé  
    return c.Status(http.StatusCreated).JSON(prod)  
}  

// Mettre à jour un produit  
func updateProduct(c *fiber.Ctx) error {  
    id := c.Params("id")  
    var prod Product  
    if err := c.BodyParser(&prod); err != nil {  
        return c.Status(http.StatusBadRequest).SendString(err.Error())  
    }  

    // Mise à jour dans la base de données  
    _, err := db.Exec(context.Background(), "UPDATE products SET name=$1, description=$2, price=$3 WHERE id=$4", prod.Name, prod.Description, prod.Price, id)  
    if err != nil {  
        return c.Status(http.StatusInternalServerError).SendString(err.Error())  
    }  
    return c.JSON(prod)  
}  

// Supprimer un produit  
func deleteProduct(c *fiber.Ctx) error {  
    id := c.Params("id")  
    _, err := db.Exec(context.Background(), "DELETE FROM products WHERE id=$1", id)  
    if err != nil {  
        return c.Status(http.StatusInternalServerError).SendString(err.Error())  
    }  
    return c.SendStatus(http.StatusNoContent)  
}
