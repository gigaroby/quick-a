package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gigaroby/quick-a/frontend/surface"
	"github.com/gigaroby/quick-a/model"
	"github.com/google/uuid"
	"honnef.co/go/js/dom"
)

type game struct {
	instructions *dom.HTMLSpanElement
	predictions  *dom.HTMLSpanElement
	countdown    *dom.HTMLSpanElement

	area *dom.HTMLDivElement
}

type wait struct {
	messages     *dom.HTMLSpanElement
	instructions *dom.HTMLSpanElement

	area *dom.HTMLDivElement
}

type final struct {
	correct *dom.HTMLSpanElement
	wrong   *dom.HTMLSpanElement
	message *dom.HTMLSpanElement

	area *dom.HTMLDivElement
}

type Session struct {
	surface *surface.S
	// channel on which draw notifications are delivered
	drawn chan struct{}

	rounds       int
	currentRound int
	roundResults []bool

	sessionID  string
	categories model.Categories

	game  game
	wait  wait
	final final
}

func swapAreas(a1, a2 *dom.HTMLDivElement) {
	a1.Style().Set("display", "")
	a2.Style().Set("display", "none")
}

func randomCategories(rounds int) (model.Categories, error) {
	res, err := http.Get("categories")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to contact backend, response code was %d %s", res.StatusCode, http.StatusText(res.StatusCode))
	}

	categories := model.Categories{}
	if err = json.NewDecoder(res.Body).Decode(&categories); err != nil {
		return nil, err
	}

	return categories, nil
}

func getPredictions(data, expectedCategory, sessionID string) (model.Predictions, error) {
	res, err := http.PostForm("classify", url.Values{
		"image":             []string{data},
		"expected_category": []string{expectedCategory},
		"session":           []string{sessionID},
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to contact backend, response code was %d %s", res.StatusCode, http.StatusText(res.StatusCode))
	}

	predictions := model.Predictions{}
	if err = json.NewDecoder(res.Body).Decode(&predictions); err != nil {
		return nil, err
	}

	return predictions, nil
}

func New(rounds int, surface *surface.S, instructions, predictions, countdown, wInstructions, wMesssages, fCorrect, fWrong, fMessage *dom.HTMLSpanElement, gameDiv, waitDiv, finalDiv *dom.HTMLDivElement) (*Session, error) {
	s := &Session{
		surface: surface,
		drawn:   make(chan struct{}),

		sessionID:    uuid.Must(uuid.NewRandom()).String(),
		rounds:       rounds,
		currentRound: 0,
		roundResults: make([]bool, rounds),

		game: game{
			instructions: instructions,
			predictions:  predictions,
			countdown:    countdown,

			area: gameDiv,
		},

		wait: wait{
			instructions: wInstructions,
			messages:     wMesssages,

			area: waitDiv,
		},

		final: final{
			correct: fCorrect,
			wrong:   fWrong,
			message: fMessage,

			area: finalDiv,
		},
	}

	surface.OnDraw = func() {
		s.drawn <- struct{}{}
	}

	cats, err := randomCategories(rounds)
	if err != nil {
		return nil, err
	}
	s.categories = cats
	s.prepareWait()

	return s, nil
}

func (s *Session) updatePredictions(preds model.Predictions) {
	if preds == nil {
		s.game.predictions.SetInnerHTML("nothing yet")
		return
	}

	buf := new(bytes.Buffer)
	for i, p := range preds {
		buf.WriteString(fmt.Sprintf("%s (%0.0f%%)", p.Category.Name, p.Confidence*100))
		if i == len(preds)-1 {
			continue
		}
		buf.WriteString(", ")
	}
	s.game.predictions.SetInnerHTML(buf.String())
}

func (s *Session) prepareWait() {
	s.wait.messages.SetInnerHTML("")
	s.wait.instructions.SetInnerHTML(s.categories[s.currentRound].Name)
	if s.currentRound == 0 {
		return
	}

	if !s.roundResults[s.currentRound-1] {
		s.wait.messages.SetInnerHTML(
			fmt.Sprintf("I didn't recognize your drawing of a %s. Better luck next time!", s.categories[s.currentRound-1].Name))
		return
	}
	s.wait.messages.SetInnerHTML(
		fmt.Sprintf("I recognized your drawing of a %s. Well done!", s.categories[s.currentRound-1].Name))
}

func (s *Session) prepareFinal() {
	var (
		correct = []string{}
		wrong   = []string{}
		message string
	)

	for i, c := range s.categories {
		if s.roundResults[i] {
			correct = append(correct, c.Name)
		} else {
			wrong = append(wrong, c.Name)
		}
	}

	percent := float32(len(correct)) / float32(s.rounds)
	switch {
	case percent == 0:
		message = "You don't even care. Do you?"
	case percent < 0.5:
		message = "I'm not even mad, just disappointed"
	case percent < 0.75:
		message = "All right, keep doing whatever it is you think you're doing."
	case percent < 0.95:
		message = "Didn't we have some fun, though?"
	default:
		message = "Unbelievable! You, &ltSubject Name Here&gt, must be the pride of &ltSubject Hometown Here&gt."
	}

	s.final.message.SetInnerHTML(message)
	s.final.correct.SetInnerHTML(strings.Join(correct, ", "))
	s.final.wrong.SetInnerHTML(strings.Join(wrong, ", "))
}

func (s *Session) NextRound() {
	currentCategory := s.categories[s.currentRound].Name
	s.game.instructions.SetInnerHTML(fmt.Sprintf("%s (%0.2d/%0.2d)", currentCategory, s.currentRound+1, s.rounds))
	s.updatePredictions(nil)
	s.game.countdown.SetInnerHTML("30")

	// show game and hide wait page
	swapAreas(s.game.area, s.wait.area)

	s.surface.Resize()
	s.surface.Clear()

	t := time.NewTicker(1 * time.Second)
	defer t.Stop()

	countdown := 30

Out:
	for {
		select {
		case <-s.drawn:
			preds, err := getPredictions(s.surface.Data(), currentCategory, s.sessionID)
			if err != nil {
				println(err)
				// TODO: show error on GUI
				continue Out
			}
			s.updatePredictions(preds)
			if preds[0].Category.Name == currentCategory {
				s.roundResults[s.currentRound] = true
				break Out
			}
		case <-t.C:
			countdown--
			s.game.countdown.SetInnerHTML(fmt.Sprintf("%0.2d", countdown))
			if countdown == 0 {
				// we did not guess correctly
				s.roundResults[s.currentRound] = false
				break Out
			}
		}
	}

	if s.currentRound >= s.rounds-1 {
		s.prepareFinal()
		swapAreas(s.final.area, s.game.area)
		return
	}

	s.currentRound++
	s.prepareWait()
	swapAreas(s.wait.area, s.game.area)
}
