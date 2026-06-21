package guitaranalysis_test

import (
	"testing"

	"github.com/wbits/guitars/internal/guitaranalysis"
	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

func TestCoverFingerprint_ChangesWhenCoverIndexChanges(t *testing.T) {
	g1 := guitarWithPictures(t, []string{"https://example.com/a.jpg", "https://example.com/b.jpg"}, 0)
	g2 := guitarWithPictures(t, []string{"https://example.com/a.jpg", "https://example.com/b.jpg"}, 1)
	if guitaranalysis.CoverFingerprintForGuitar(g1) == guitaranalysis.CoverFingerprintForGuitar(g2) {
		t.Fatal("expected different fingerprint when cover index changes")
	}
}

func TestCoverFingerprint_UnchangedWhenOtherPicturesChange(t *testing.T) {
	g1 := guitarWithPictures(t, []string{"https://example.com/a.jpg", "https://example.com/b.jpg"}, 0)
	g2 := guitarWithPictures(t, []string{"https://example.com/a.jpg", "https://example.com/c.jpg"}, 0)
	if guitaranalysis.CoverFingerprintForGuitar(g1) != guitaranalysis.CoverFingerprintForGuitar(g2) {
		t.Fatal("expected same fingerprint when cover picture is unchanged")
	}
}

func TestCoverPictureURL_UsesCoverPictureIndex(t *testing.T) {
	g := guitarWithPictures(t, []string{"https://example.com/a.jpg", "https://example.com/b.jpg"}, 1)
	if got := guitaranalysis.CoverPictureURL(g); got != "https://example.com/b.jpg" {
		t.Fatalf("cover url: %q", got)
	}
}

func guitarWithPictures(t *testing.T, pictures []string, coverIndex int) *domain.Guitar {
	t.Helper()
	price, err := domain.NewMoney(100000, domain.EUR)
	if err != nil {
		t.Fatal(err)
	}
	g, err := domain.NewGuitar(domain.GuitarProps{
		ID: "g1", Owner: "owner-1", Brand: "Fender", TypeName: "Strat",
		BuildYear: 1996, Price: price, Pictures: pictures, CoverPictureIndex: coverIndex,
	})
	if err != nil {
		t.Fatal(err)
	}
	return g
}
