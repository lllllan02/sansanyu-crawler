package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExam(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		exam, err := client.GetExam("http://www.sansanyu.com/index.php?learn-app-shijuan-detail&examid=462")
		assert.NoError(t, err)
		assert.NotEmpty(t, exam.Title)
		assert.NotEmpty(t, exam.Problems)

		fmt.Printf("title: %v\n", exam.Title)
		for _, problem := range exam.Problems {
			fmt.Printf("problem: %+v\n", problem)
		}
	})
}
