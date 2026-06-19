import { create } from "zustand";

interface ChatSessionStore {
  selectedSessionId: string;
  selectedSessionTitle: string;
  setSelectedSession: (sessionId: string, title?: string) => void;
  sidebarRefreshVersion: number;
  setSelectedSessionId: (sessionId: string) => void;
  clearSelectedSessionId: () => void;
  requestSidebarRefresh: () => void;
}

/**
 * useChatSessionStore 管理 Chat 侧边栏和正文共享的会话状态。
 */
export const useChatSessionStore = create<ChatSessionStore>((set) => ({
  selectedSessionId: "",
  selectedSessionTitle: "",
  sidebarRefreshVersion: 0,

  setSelectedSession: (sessionId: string, title?: string) => {
    set({
      selectedSessionId: sessionId,
      selectedSessionTitle: title?.trim() || "新对话",
    });
  },

  setSelectedSessionId: (sessionId: string) => {
    set((state) => ({
      selectedSessionId: sessionId,
      selectedSessionTitle:
        state.selectedSessionId === sessionId
          ? state.selectedSessionTitle
          : "新对话",
    }));
  },

  clearSelectedSessionId: () => {
    set({ selectedSessionId: "", selectedSessionTitle: "" });
  },

  requestSidebarRefresh: () => {
    set((state) => ({
      sidebarRefreshVersion: state.sidebarRefreshVersion + 1,
    }));
  },
}));
