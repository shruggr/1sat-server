package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/libsv/go-bt"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/shruggr/1sat-indexer/lib"
	// _ "github.com/shruggr/bsv-ord-indexer/server/docs"
	// swaggerFiles "github.com/swaggo/files"
	// ginSwagger "github.com/swaggo/gin-swagger"
)

var db *sql.DB
var index []byte

func init() {
	godotenv.Load()

	var err error
	db, err = sql.Open("postgres", os.Getenv("POSTGRES"))
	if err != nil {
		log.Fatal(err)
	}

	err = lib.Initialize(db)
	if err != nil {
		log.Fatal(err)
	}

	index, err = os.ReadFile("./static/index.html")
}

func main() {
	r := gin.Default()
	// url := ginSwagger.URL("/swagger/doc.json")
	// r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		// AllowOriginFunc: func(origin string) bool {
		// 	return origin == "https://github.com"
		// },
		MaxAge: 12 * time.Hour,
	}))

	r.GET("/api/utxos/address/:address", func(c *gin.Context) {
		script, err := bscript.NewP2PKHFromAddress(c.Param("address"))
		if err != nil {
			handleError(c, &lib.HttpError{
				StatusCode: http.StatusBadRequest,
				Err:        err,
			})
		}
		hash := sha256.Sum256(*script)
		lock := bt.ReverseBytes(hash[:])
		utxos, err := lib.LoadUtxos(lock)
		if err != nil {
			handleError(c, err)
		}

		c.Header("cache-control", "no-cache,no-store,must-revalidate")
		c.JSON(http.StatusOK, utxos)
	})

	r.GET("/api/utxos/lock/:lock", func(c *gin.Context) {
		lock, err := hex.DecodeString(c.Param("lock"))
		if err != nil {
			handleError(c, &lib.HttpError{
				StatusCode: http.StatusBadRequest,
				Err:        err,
			})
		}
		utxos, err := lib.LoadUtxos(lock)
		if err != nil {
			handleError(c, err)
		}

		c.Header("cache-control", "no-cache,no-store,must-revalidate")
		c.JSON(http.StatusOK, utxos)
	})

	r.GET("/api/inscriptions/origin/:origin", func(c *gin.Context) {
		origin, err := lib.NewOriginFromString(c.Param("origin"))
		if err != nil {
			handleError(c, err)
			return
		}
		im, err := lib.LoadInscriptions(origin)
		if err != nil {
			handleError(c, err)
			return
		}

		c.Header("cache-control", "max-age=604800,immutable")
		c.JSON(http.StatusOK, im)
	})

	r.GET("/api/inscriptions/txid/:txid", func(c *gin.Context) {
		txid, err := hex.DecodeString(c.Param("txid"))
		if err != nil {
			handleError(c, err)
			return
		}

		im, err := lib.LoadInscriptionsByTxID(txid)
		if err != nil {
			handleError(c, err)
			return
		}

		c.Header("cache-control", "max-age=604800,immutable")
		c.JSON(http.StatusOK, im)
	})

	r.GET("/api/inscriptions/count", func(c *gin.Context) {
		count, err := lib.GetInscriptionCount()
		if err != nil {
			handleError(c, err)
			return
		}
		// c.Header("cache-control", "no-cache,no-store,must-revalidate")
		c.JSON(http.StatusOK, gin.H{"count": count})
	})

	r.GET("/api/inscriptions/:id", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			handleError(c, err)
			return
		}
		im, err := lib.LoadInscriptionById(id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.Header("cache-control", "max-age=604800,immutable")
		c.JSON(http.StatusOK, im)
	})

	r.GET("/api/files/inscriptions/:origin", func(c *gin.Context) {
		origin, err := lib.NewOriginFromString(c.Param("origin"))
		if err != nil {
			handleError(c, err)
			return
		}
		ins, err := lib.LoadInscriptionFile(origin)
		if err != nil {
			handleError(c, err)
			return
		}
		c.Header("cache-control", "max-age=604800,immutable")
		c.Data(http.StatusOK, ins.Type, ins.Body)
	})

	r.GET("/api/handle/*pattern", func(c *gin.Context) {
		// c.String(http.StatusOK, "OK")
		fmt.Println(c.Param("pattern"))
		if strings.HasPrefix(c.Param("pattern"), "/web+onesatorg") {
			parts := strings.Split(c.Param("pattern"), "/")
			c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/api/files/inscriptions/%s", parts[len(parts)-1]))
			return
		}
		c.String(http.StatusNotFound, "Not found")

	})
	r.NoRoute(func(c *gin.Context) {
		txtRecs, err := net.LookupTXT(c.Request.Host)
		if err != nil {
			handleError(c, err)
			return
		}
		for _, txt := range txtRecs {
			if !strings.HasPrefix(txt, "1sat-origin=") {
				origin, err := lib.NewOriginFromString(strings.TrimPrefix(txt, "1sat-origin="))
				if err != nil {
					handleError(c, err)
					return
				}
				ins, err := lib.LoadInscriptionFile(origin)
				if err != nil {
					handleError(c, err)
					return
				}
				c.Header("cache-control", "max-age=604800,immutable")
				c.Data(http.StatusOK, ins.Type, ins.Body)
				return
			}
		}
		c.String(200, "OK")
	})
	listen := os.Getenv("LISTEN")
	if listen == "" {
		listen = "0.0.0.0:8080"
	}
	r.Run(listen) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func handleError(c *gin.Context, err error) {
	if httpErr, ok := err.(*lib.HttpError); ok {
		c.String(httpErr.StatusCode, "%v", httpErr.Err)
	} else {
		c.String(http.StatusInternalServerError, "%v", err)
	}
}
