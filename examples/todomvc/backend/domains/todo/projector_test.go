// Copyright (c) 2017 - The Event Horizon authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package todo

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/kr/pretty"

	eh "github.com/reidlai/eventhorizon"
	"github.com/reidlai/eventhorizon/uuid"
)

func TestProjector(t *testing.T) {
	TimeNow = func() time.Time {
		return time.Date(2017, time.July, 10, 23, 0, 0, 0, time.Local)
	}

	id := uuid.New()
	cases := map[string]struct {
		model         eh.Entity
		event         eh.Event
		expectedModel eh.Entity
		expectedErr   error
	}{
		"unhandeled event": {
			&TodoList{},
			eh.NewEvent(eh.EventType("unknown"), nil, TimeNow(),
				eh.ForAggregate(AggregateType, id, 1),
			),
			&TodoList{},
			errors.New("could not project event: unknown"),
		},
		"created": {
			&TodoList{},
			eh.NewEvent(Created, nil, TimeNow(),
				eh.ForAggregate(AggregateType, id, 1),
			),
			&TodoList{
				ID:        id,
				Version:   1,
				Items:     []*TodoItem{},
				CreatedAt: TimeNow(),
				UpdatedAt: TimeNow(),
			},
			nil,
		},
		"deleted": {
			&TodoList{},
			eh.NewEvent(Deleted, nil, TimeNow(),
				eh.ForAggregate(AggregateType, id, 1),
			),
			nil,
			nil,
		},
		"item added": {
			&TodoList{
				ID:        id,
				Version:   1,
				Items:     []*TodoItem{},
				CreatedAt: TimeNow(),
			},
			eh.NewEvent(ItemAdded, &ItemAddedData{
				ItemID:      1,
				Description: "desc 1",
			}, TimeNow(),
				eh.ForAggregate(AggregateType, id, 1),
			),
			&TodoList{
				ID:      id,
				Version: 2,
				Items: []*TodoItem{
					{
						ID:          1,
						Description: "desc 1",
						Completed:   false,
					},
				},
				CreatedAt: TimeNow(),
				UpdatedAt: TimeNow(),
			},
			nil,
		},
		"item removed": {
			&TodoList{
				ID:      id,
				Version: 1,
				Items: []*TodoItem{
					{
						ID:          1,
						Description: "desc 1",
						Completed:   false,
					},
					{
						ID:          2,
						Description: "desc 2",
						Completed:   false,
					},
				},
				CreatedAt: TimeNow(),
			},
			eh.NewEvent(ItemRemoved, &ItemRemovedData{
				ItemID: 2,
			}, TimeNow(),
				eh.ForAggregate(AggregateType, id, 1),
			),
			&TodoList{
				ID:      id,
				Version: 2,
				Items: []*TodoItem{
					{
						ID:          1,
						Description: "desc 1",
						Completed:   false,
					},
				},
				CreatedAt: TimeNow(),
				UpdatedAt: TimeNow(),
			},
			nil,
		},
		"item removed (last)": {
			&TodoList{
				ID:      id,
				Version: 1,
				Items: []*TodoItem{
					{
						ID:          1,
						Description: "desc 1",
						Completed:   false,
					},
				},
				CreatedAt: TimeNow(),
			},
			eh.NewEvent(ItemRemoved, &ItemRemovedData{
				ItemID: 1,
			}, TimeNow(),
				eh.ForAggregate(AggregateType, id, 1),
			),
			&TodoList{
				ID:        id,
				Version:   2,
				Items:     []*TodoItem{},
				CreatedAt: TimeNow(),
				UpdatedAt: TimeNow(),
			},
			nil,
		},
		"item description set": {
			&TodoList{
				ID:      id,
				Version: 1,
				Items: []*TodoItem{
					{
						ID:          1,
						Description: "desc 1",
						Completed:   false,
					},
					{
						ID:          2,
						Description: "desc 2",
						Completed:   false,
					},
				},
				CreatedAt: TimeNow(),
			},
			eh.NewEvent(ItemDescriptionSet, &ItemDescriptionSetData{
				ItemID:      2,
				Description: "new desc",
			}, TimeNow(),
				eh.ForAggregate(AggregateType, id, 1),
			),
			&TodoList{
				ID:      id,
				Version: 2,
				Items: []*TodoItem{
					{
						ID:          1,
						Description: "desc 1",
						Completed:   false,
					},
					{
						ID:          2,
						Description: "new desc",
						Completed:   false,
					},
				},
				CreatedAt: TimeNow(),
				UpdatedAt: TimeNow(),
			},
			nil,
		},
		"item checked": {
			&TodoList{
				ID:      id,
				Version: 1,
				Items: []*TodoItem{
					{
						ID:          1,
						Description: "desc 1",
						Completed:   false,
					},
					{
						ID:          2,
						Description: "desc 2",
						Completed:   false,
					},
				},
				CreatedAt: TimeNow(),
			},
			eh.NewEvent(ItemChecked, &ItemCheckedData{
				ItemID:  2,
				Checked: true,
			}, TimeNow(),
				eh.ForAggregate(AggregateType, id, 1),
			),
			&TodoList{
				ID:      id,
				Version: 2,
				Items: []*TodoItem{
					{
						ID:          1,
						Description: "desc 1",
						Completed:   false,
					},
					{
						ID:          2,
						Description: "desc 2",
						Completed:   true,
					},
				},
				CreatedAt: TimeNow(),
				UpdatedAt: TimeNow(),
			},
			nil,
		},
	}

	for name, tc := range cases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			projector := &Projector{}
			model, err := projector.Project(context.Background(), tc.event, tc.model)
			if (err != nil && tc.expectedErr == nil) ||
				(err == nil && tc.expectedErr != nil) ||
				(err != nil && tc.expectedErr != nil && err.Error() != tc.expectedErr.Error()) {
				t.Errorf("test case '%s': incorrect error", name)
				t.Log("exp:", tc.expectedErr)
				t.Log("got:", err)
			}
			if !reflect.DeepEqual(model, tc.expectedModel) {
				t.Errorf("test case '%s': incorrect aggregate", name)
				t.Log("exp:\n", pretty.Sprint(tc.expectedModel))
				t.Log("got:\n", pretty.Sprint(model))
			}
		})
	}
}
