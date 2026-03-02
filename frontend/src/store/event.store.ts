// Copyright 2026 chenyang
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { create } from "zustand";
import { Events } from "@wailsio/runtime";
import { StoreMethods } from "./common";
import { GitStatusChangedEvent } from "@wails/types/models";
import { EventType } from "@wails/events/models";

export const defaultTrackedEvents = [
  EventType.EventTypeGitStatusChanged,
] as const;

// 注意：该类型元组与 defaultTrackedEvents 必须保持“同下标同语义”映射关系。
export type defaultTrackedEventsType = [GitStatusChangedEvent];

type TupleKeys<T extends readonly unknown[]> = Exclude<keyof T, keyof any[]>;

type DefaultEventPayloadMap = {
  [I in TupleKeys<
    typeof defaultTrackedEvents
  > as (typeof defaultTrackedEvents)[I] & string]: defaultTrackedEventsType[I];
};

export type DefaultEventName = keyof DefaultEventPayloadMap & string;
export type EventName = DefaultEventName;

type EventPayloadByName<K extends EventName> = K extends DefaultEventName
  ? DefaultEventPayloadMap[K]
  : never;

interface EventPayloadEntry<T = unknown> {
  data: T;
  timestamp: number;
}

type DefaultLatestEvents = {
  [K in DefaultEventName]?: EventPayloadEntry<DefaultEventPayloadMap[K]>;
};

interface EventStoreState {
  initialized: boolean;
  trackedEvents: DefaultEventName[];
  latestEvents: DefaultLatestEvents;
  listeners: Map<DefaultEventName, () => void>;

  initialize: (eventNames?: DefaultEventName[]) => void;
  dispose: () => void;
  setEvent: <K extends EventName>(
    eventName: K,
    payload: EventPayloadByName<K>,
  ) => void;
  getEvent: <K extends EventName>(
    eventName: K,
  ) => EventPayloadEntry<EventPayloadByName<K>> | undefined;
  clearEvent: (eventName: EventName) => void;
  clearAll: () => void;
}

function toPayload(raw: unknown): unknown {
  if (
    raw &&
    typeof raw === "object" &&
    "data" in (raw as Record<string, unknown>)
  ) {
    return (raw as { data: unknown }).data;
  }
  return raw;
}

export const useEventStore = create<EventStoreState>((set, get) => ({
  initialized: false,
  trackedEvents: [],
  latestEvents: {},
  listeners: new Map(),

  initialize: (eventNames = [...defaultTrackedEvents]) => {
    const state = get();
    if (state.initialized) return;

    const listeners = new Map<DefaultEventName, () => void>();

    for (const eventName of eventNames) {
      const unbind = Events.On(eventName, (event: unknown) => {
        get().setEvent(
          eventName,
          toPayload(event) as EventPayloadByName<typeof eventName>,
        );
      });
      listeners.set(eventName, unbind);
    }

    set({
      initialized: true,
      trackedEvents: [...eventNames] as DefaultEventName[],
      listeners,
    });
  },

  dispose: () => {
    const state = get();
    for (const unbind of state.listeners.values()) {
      unbind();
    }

    set({
      initialized: false,
      trackedEvents: [],
      listeners: new Map<DefaultEventName, () => void>(),
      latestEvents: {},
    });
  },

  setEvent: (eventName, payload) => {
    set((state) => ({
      latestEvents: {
        ...state.latestEvents,
        [eventName]: {
          data: payload,
          timestamp: Date.now(),
        },
      } as DefaultLatestEvents,
    }));
  },

  getEvent: (eventName) =>
    get().latestEvents[eventName] as
      | EventPayloadEntry<EventPayloadByName<typeof eventName>>
      | undefined,

  clearEvent: (eventName) => {
    set((state) => {
      const next = { ...state.latestEvents };
      delete next[eventName];
      return { latestEvents: next };
    });
  },

  clearAll: () => {
    set({ latestEvents: {} });
  },
}));

export const eventStoreMethods = StoreMethods(useEventStore);
