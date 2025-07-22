package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExam(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		title, problems, err := client.GetExam("http://www.sansanyu.com/index.php?learn-app-shijuan-detail&examid=462")
		assert.NoError(t, err)
		assert.NotEmpty(t, title)
		assert.NotEmpty(t, problems)

		fmt.Printf("title: %v\n", title)
		for _, problem := range problems {
			fmt.Printf("problem: %+v\n", problem)
		}
	})
}
