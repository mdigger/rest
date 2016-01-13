package rest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContext(t *testing.T) {
	r, err := http.NewRequest("GET", "/test?akdjf#sdf", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	c := newContext(w, r)
	header := c.Header()
	header.Set("test", "32")
	fmt.Println(c.Header().Get("test"))
	c.close()

}
