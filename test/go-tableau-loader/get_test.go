package loader_test

import (
	"errors"
	"testing"

	"github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader"
)

// ---- ItemConf: 1st-level int-keyed map getter ----

func Test_ItemConf_Get1(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetItemConf()
	if conf == nil {
		t.Fatal("ItemConf is nil")
	}

	// found
	item, err := conf.Get1(1)
	if err != nil {
		t.Fatalf("Get1(1) unexpected error: %v", err)
	}
	if item.GetName() != "apple" {
		t.Errorf("Get1(1): expected name=apple, got %s", item.GetName())
	}
	item, err = conf.Get1(0)
	if err != nil {
		t.Fatalf("Get1(0) unexpected error: %v", err)
	}
	if item.GetName() != "coin1" {
		t.Errorf("Get1(0): expected name=coin1, got %s", item.GetName())
	}

	// not found
	if _, err := conf.Get1(12345); !errors.Is(err, loader.ErrNotFound) {
		t.Fatalf("Get1(12345): expected ErrNotFound, got: %v", err)
	}
}

// ---- ActivityConf: nested map getters with a 64-bit outer key ----

func Test_ActivityConf_Get(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetActivityConf()
	if conf == nil {
		t.Fatal("ActivityConf is nil")
	}

	activity, err := conf.Get1(100001)
	if err != nil {
		t.Fatalf("Get1(100001) unexpected error: %v", err)
	}
	if activity.GetActivityName() != "活动1" {
		t.Errorf("Get1(100001): expected activityName=活动1, got %s", activity.GetActivityName())
	}

	chapter, err := conf.Get2(100001, 1)
	if err != nil {
		t.Fatalf("Get2(100001, 1) unexpected error: %v", err)
	}
	if chapter.GetChapterName() != "签到活动章1" {
		t.Errorf("Get2(100001, 1): expected chapterName=签到活动章1, got %s", chapter.GetChapterName())
	}

	section, err := conf.Get3(100001, 1, 2)
	if err != nil {
		t.Fatalf("Get3(100001, 1, 2) unexpected error: %v", err)
	}
	if section.GetSectionId() != 2 {
		t.Errorf("Get3(100001, 1, 2): expected sectionId=2, got %d", section.GetSectionId())
	}

	// sectionRankMap["2001"] === 2
	rank, err := conf.Get4(100001, 1, 2, 2001)
	if err != nil {
		t.Fatalf("Get4(100001, 1, 2, 2001) unexpected error: %v", err)
	}
	if rank != 2 {
		t.Errorf("Get4(100001, 1, 2, 2001): expected 2, got %d", rank)
	}
}

func Test_ActivityConf_NotFound(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetActivityConf()
	if conf == nil {
		t.Fatal("ActivityConf is nil")
	}
	if _, err := conf.Get3(100001, 1, 999); !errors.Is(err, loader.ErrNotFound) {
		t.Fatalf("Get3(100001, 1, 999): expected ErrNotFound, got: %v", err)
	}
}

// ---- ThemeConf: string-keyed map getter ----

func Test_ThemeConf_Get1(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetThemeConf()
	if conf == nil {
		t.Fatal("ThemeConf is nil")
	}
	if len(conf.Data().GetThemeMap()) == 0 {
		t.Fatal("ThemeConf: themeMap is empty")
	}
}

// ---- HeroBaseConf ----

func Test_HeroBaseConf(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetHeroBaseConf()
	if conf == nil {
		t.Fatal("HeroBaseConf is nil")
	}
	if len(conf.Data().GetHeroMap()) == 0 {
		t.Fatal("HeroBaseConf: heroMap is empty")
	}
}

// ---- FruitConf: leveled point getters ----

func Test_FruitConf_Get1(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruitConf()

	// found
	fruit, err := conf.Get1(fruitTypeApple)
	if err != nil {
		t.Fatalf("Get1(%d) unexpected error: %v", fruitTypeApple, err)
	}
	if fruit == nil {
		t.Fatal("Get1: returned nil fruit")
	}

	// not found
	if _, err := conf.Get1(999); !errors.Is(err, loader.ErrNotFound) {
		t.Fatalf("Get1(999): expected ErrNotFound, got: %v", err)
	}
}

func Test_FruitConf_Get2(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruitConf()

	// found: APPLE -> item 1001
	item, err := conf.Get2(fruitTypeApple, 1001)
	if err != nil {
		t.Fatalf("Get2(%d, 1001) unexpected error: %v", fruitTypeApple, err)
	}
	if item.GetId() != 1001 {
		t.Errorf("Get2: expected id=1001, got %d", item.GetId())
	}
	if item.GetPrice() != 10 {
		t.Errorf("Get2: expected price=10, got %d", item.GetPrice())
	}

	// not found: wrong item id
	if _, err := conf.Get2(fruitTypeApple, 9999); !errors.Is(err, loader.ErrNotFound) {
		t.Fatalf("Get2(apple, 9999): expected ErrNotFound, got: %v", err)
	}

	// not found: wrong fruitType
	if _, err := conf.Get2(999, 1001); !errors.Is(err, loader.ErrNotFound) {
		t.Fatalf("Get2(999, 1001): expected ErrNotFound, got: %v", err)
	}
}

func Test_Fruit6Conf_Get1(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruit6Conf()

	// found
	fruit, err := conf.Get1(fruitTypeApple)
	if err != nil {
		t.Fatalf("Get1(%d) unexpected error: %v", fruitTypeApple, err)
	}
	if fruit == nil {
		t.Fatal("Get1: returned nil fruit")
	}

	// not found
	if _, err := conf.Get1(999); !errors.Is(err, loader.ErrNotFound) {
		t.Fatalf("Get1(999): expected ErrNotFound, got: %v", err)
	}
}
