package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSubject(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		links, next, err := client.GetSubject("http://www.sansanyu.com/index.php?learn-app-shijuan&subjectid=18&page=1")
		assert.NoError(t, err)
		assert.Len(t, links, 10)
		assert.NotEmpty(t, next)

		for _, link := range links {
			fmt.Printf("link: %+v\n", link)
		}
	})

	t.Run("2", func(t *testing.T) {
		links, next, err := client.GetSubject("http://www.sansanyu.com/index.php?learn-app-shijuan&subjectid=18&page=2")
		assert.NoError(t, err)
		assert.Len(t, links, 8)
		assert.Empty(t, next)
	})
}
