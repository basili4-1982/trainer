package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"trainer/pkg/task"
)

func main() {
	g := gin.Default()

	g.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})

	g.POST("/task", func(c *gin.Context) {
		t := task.Task{}
		if err := c.ShouldBind(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res := task.TaskRunner(c, t)
		c.JSON(http.StatusOK, gin.H{"result": gin.H{
			"task_name": res.TaskName,
			"output":    res.Output,
			"error":     res.Error,
			"success":   res.Success,
		},
		})
	})

	err := g.Run(":8080")
	if err != nil {
		log.Fatalln(errors.Wrap(err, "failed to run gin server"))
	}
}
