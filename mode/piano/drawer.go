package piano

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hndada/gosu"
	"github.com/hndada/gosu/draws"
)

type StageDrawer struct {
	FieldSprite draws.Sprite
	HintSprite  draws.Sprite
}

// Todo: might add some effect on StageDrawer
func (d StageDrawer) Draw(screen *ebiten.Image) {
	d.FieldSprite.Draw(screen, nil)
	d.HintSprite.Draw(screen, nil)
}

// Bars are fixed. Lane itself moves, all bars move as same amount.
type BarDrawer struct {
	Cursor   float64
	Farthest *Bar
	Nearest  *Bar
	Sprite   draws.Sprite
}

func (d *BarDrawer) Update(cursor float64) {
	d.Cursor = cursor
	// When Farthest's prevs are still out of screen due to speed change.
	for d.Farthest.Prev != nil &&
		d.Farthest.Prev.Position-d.Cursor > maxPosition {
		d.Farthest = d.Farthest.Prev
	}
	// When Farthest is in screen, next note goes fetched if possible.
	for d.Farthest.Next != nil &&
		d.Farthest.Position-d.Cursor <= maxPosition {
		d.Farthest = d.Farthest.Next
	}
	// When Nearest is still in screen due to speed change.
	for d.Nearest.Prev != nil &&
		d.Nearest.Position-d.Cursor > minPosition {
		d.Nearest = d.Nearest.Prev
	}
	// When Nearest's next is still out of screen, next note goes fetched.
	for d.Nearest.Next != nil &&
		d.Nearest.Next.Position-d.Cursor <= minPosition {
		d.Nearest = d.Nearest.Next
	}
}

func (d BarDrawer) Draw(screen *ebiten.Image) {
	if d.Farthest == nil || d.Nearest == nil {
		return
	}
	for b := d.Farthest; b != d.Nearest.Prev; b = b.Prev {
		sprite := d.Sprite
		pos := b.Position - d.Cursor
		sprite.Move(0, -pos)
		sprite.Draw(screen, nil)
	}
}

// Notes are fixed. Lane itself moves, all notes move same amount.
type NoteDrawer struct {
	Cursor   float64
	Farthest *Note
	Nearest  *Note
	Sprites  [4]draws.Sprite
}

// Farthest and Nearest are borders of displaying notes.
// All in-screen notes are confirmed to be drawn when drawing from Farthest to Nearest.
func (d *NoteDrawer) Update(cursor float64) {
	d.Cursor = cursor
	if d.Farthest == nil || d.Nearest == nil {
		return
	}
	// When Farthest's prevs are still out of screen due to speed change.
	for d.Farthest.Prev != nil &&
		d.Farthest.Prev.Position-d.Cursor > maxPosition {
		d.Farthest = d.Farthest.Prev
	}
	// When Farthest is in screen, next note goes fetched if possible.
	for d.Farthest.Next != nil &&
		d.Farthest.Position-d.Cursor <= maxPosition {
		d.Farthest = d.Farthest.Next
	}
	// When Nearest is still in screen due to speed change.
	for d.Nearest.Prev != nil &&
		d.Nearest.Position-d.Cursor > minPosition {
		d.Nearest = d.Nearest.Prev
	}
	// When Nearest's next is still out of screen, next note goes fetched.
	for d.Nearest.Next != nil &&
		d.Nearest.Next.Position-d.Cursor <= minPosition {
		d.Nearest = d.Nearest.Next
	}
}

// Draw from farthest to nearest to make nearer notes priorly exposed.
func (d NoteDrawer) Draw(screen *ebiten.Image) {
	if d.Farthest == nil || d.Nearest == nil {
		return
	}
	for n := d.Farthest; n != nil && n != d.Nearest.Prev; n = n.Prev {
		if n.Type == Tail {
			d.DrawLongBody(screen, n)
		}
		sprite := d.Sprites[n.Type]
		pos := n.Position - d.Cursor
		sprite.Move(0, -pos)
		op := &ebiten.DrawImageOptions{}
		if n.Marked {
			op.ColorM.ChangeHSV(0, 0.3, 0.3)
		}
		sprite.Draw(screen, op)
	}
}

// DrawLongBody draws scaled, corresponding sub-image of Body sprite.
func (d NoteDrawer) DrawLongBody(screen *ebiten.Image, tail *Note) {
	head := tail.Prev
	body := d.Sprites[Body]
	length := tail.Position - head.Position
	length -= -bodyLoss
	ratio := length / body.H()
	op := &ebiten.DrawImageOptions{}
	if ReverseBody {
		body.SetScaleXY(1, -ratio, ebiten.FilterLinear)
	} else {
		body.SetScaleXY(1, ratio, ebiten.FilterLinear)
	}
	if tail.Marked {
		op.ColorM.ChangeHSV(0, 0.3, 0.3)
	}
	ty := head.Position - d.Cursor
	body.Move(0, -ty)
	body.Draw(screen, op)
}

// KeyDrawer draws KeyDownSprite at least for 30ms, KeyUpSprite otherwise.
// KeyDrawer uses MinCountdown instead of MaxCountdown.
type KeyDrawer struct {
	MinCountdown   int
	Countdowns     []int
	KeyUpSprites   []draws.Sprite
	KeyDownSprites []draws.Sprite
	lastPressed    []bool
	pressed        []bool
}

func (d *KeyDrawer) Update(lastPressed, pressed []bool) {
	d.lastPressed = lastPressed
	d.pressed = pressed
	for k, countdown := range d.Countdowns {
		if countdown > 0 {
			d.Countdowns[k]--
		}
	}
	for k, now := range d.pressed {
		last := d.lastPressed[k]
		if !last && now {
			d.Countdowns[k] = d.MinCountdown
		}
	}
}
func (d KeyDrawer) Draw(screen *ebiten.Image) {
	for k, p := range d.pressed {
		if p || d.Countdowns[k] > 0 {
			d.KeyDownSprites[k].Draw(screen, nil)
		} else {
			d.KeyUpSprites[k].Draw(screen, nil)
		}
	}
}

type JudgmentDrawer struct {
	draws.BaseDrawer
	Sprites  []draws.Sprite
	Judgment gosu.Judgment
}

func NewJudgmentDrawer() (d JudgmentDrawer) {
	return JudgmentDrawer{
		BaseDrawer: draws.BaseDrawer{
			MaxCountdown: gosu.TimeToTick(600),
		},
		Sprites: GeneralSkin.JudgmentSprites,
	}
}
func (d *JudgmentDrawer) Update(worst gosu.Judgment) {
	if d.Countdown <= 0 {
		d.Judgment = gosu.Judgment{}
	} else {
		d.Countdown--
	}
	if worst.Valid() {
		d.Judgment = worst
		d.Countdown = d.MaxCountdown
	}
}

func (d JudgmentDrawer) Draw(screen *ebiten.Image) {
	if d.Countdown <= 0 || d.Judgment.Window == 0 {
		return
	}
	var sprite draws.Sprite
	for i, j := range Judgments {
		if j.Window == d.Judgment.Window {
			sprite = d.Sprites[i]
			break
		}
	}
	age := d.Age()
	ratio := 1.0
	switch {
	case age < 0.1:
		ratio = 1.15 * (1 + age)
	case age >= 0.1 && age < 0.2:
		ratio = 1.15 * (1.2 - age)
	case age > 0.9:
		ratio = 1 - 1.15*(age-0.9)
	}
	sprite.SetScale(ratio)
	sprite.Draw(screen, nil)
}
