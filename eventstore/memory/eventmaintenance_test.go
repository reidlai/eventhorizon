// Copyright (c) 2014 - The Event Horizon authors.
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

package memory

import (
	"context"
	"testing"

	"github.com/reidlai/eventhorizon/eventstore"
)

func TestEventStoreMaintenance(t *testing.T) {
	store, err := NewEventStore()
	if err != nil {
		t.Fatal("there should be no error:", err)
	}

	if store == nil {
		t.Fatal("there should be a store")
	}

	eventstore.MaintenanceAcceptanceTest(t, store, store, context.Background())
}
