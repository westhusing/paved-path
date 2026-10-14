package jobs

import "testing"

func TestProcess(t *testing.T) {
	t.Run("valid job", func(t *testing.T) {
		res, err := Process(Job{ID: "job-1", Payload: "x"})
		if err != nil {
			t.Fatal(err)
		}
		if res.Status != "ok" || !res.Handled {
			t.Fatalf("got %+v", res)
		}
	})
	t.Run("empty id rejected", func(t *testing.T) {
		if _, err := Process(Job{}); err != ErrEmptyID {
			t.Fatalf("err = %v, want ErrEmptyID", err)
		}
	})
}
