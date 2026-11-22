import { PayloadAction, createSlice } from '@reduxjs/toolkit';
import { AiChatMessage } from '@graphql';
import { v4 as uuidv4 } from 'uuid';

export type IChatMessage = AiChatMessage & {
    isUserInput?: boolean;
    isStreaming?: boolean;
    id?: number;
};

export type ChatSession = {
    id: string;
    name: string;
    messages: IChatMessage[];
    createdAt: Date;
};

export type IChatState = {

    chats: IChatMessage[];

    sessions: ChatSession[];
    activeSessionId: string | null;
}

const initialState: IChatState = {
    chats: [],
    sessions: [],
    activeSessionId: null,
}

export const houdiniSlice = createSlice({
  name: 'chats',
  initialState,
  reducers: {

    addChatMessage: (state, action: PayloadAction<IChatMessage>) => {

        if (state.sessions.length > 0 && state.activeSessionId) {
            const session = state.sessions.find(s => s.id === state.activeSessionId);
            if (session) {
                session.messages.push(action.payload);
            }
        } else {

            state.chats.push(action.payload);
        }
    },
    updateChatMessage: (state, action: PayloadAction<{ id: number; Text: string }>) => {

        if (state.sessions.length > 0 && state.activeSessionId) {
            const session = state.sessions.find(s => s.id === state.activeSessionId);
            if (session) {
                const message = session.messages.find(chat => chat.id === action.payload.id);
                if (message) {
                    message.Text = action.payload.Text;
                }
            }
        } else {

            const message = state.chats.find(chat => chat.id === action.payload.id);
            if (message) {
                message.Text = action.payload.Text;
            }
        }
    },
    completeStreamingMessage: (state, action: PayloadAction<{ id: number; message: Partial<IChatMessage> }>) => {

        if (state.sessions.length > 0 && state.activeSessionId) {
            const session = state.sessions.find(s => s.id === state.activeSessionId);
            if (session) {
                const message = session.messages.find(chat => chat.id === action.payload.id);
                if (message) {
                    Object.assign(message, action.payload.message);
                    message.isStreaming = false;
                }
            }
        } else {

            const message = state.chats.find(chat => chat.id === action.payload.id);
            if (message) {
                Object.assign(message, action.payload.message);
                message.isStreaming = false;
            }
        }
    },
    removeChatMessage: (state, action: PayloadAction<number>) => {

        if (state.sessions.length > 0 && state.activeSessionId) {
            const session = state.sessions.find(s => s.id === state.activeSessionId);
            if (session) {
                session.messages = session.messages.filter(chat => chat.id !== action.payload);
            }
        } else {

            state.chats = state.chats.filter(chat => chat.id !== action.payload);
        }
    },
    clear: (state) => {

        if (state.sessions.length > 0 && state.activeSessionId) {
            const session = state.sessions.find(s => s.id === state.activeSessionId);
            if (session) {
                session.messages = [];
            }
        } else {

            state.chats = [];
        }
    },

    initializeChatSessions: (state) => {
        if (state.sessions.length === 0) {
            const newId = uuidv4();
            const newSession: ChatSession = {
                id: newId,
                name: "Chat 1",
                messages: [],
                createdAt: new Date()
            };
            state.sessions = [newSession];
            state.activeSessionId = newId;
        }

        if (!state.activeSessionId && state.sessions.length > 0) {
            state.activeSessionId = state.sessions[0].id;
        }
    },
    addChatSession: (state, action: PayloadAction<{ name?: string }>) => {
        const newId = uuidv4();
        const newSession: ChatSession = {
            id: newId,
            name: action.payload.name || `Chat ${state.sessions.length + 1}`,
            messages: [],
            createdAt: new Date()
        };

        state.sessions.unshift(newSession);
        state.activeSessionId = newId;
    },
    deleteChatSession: (state, action: PayloadAction<{ sessionId: string }>) => {
        if (state.sessions.length <= 1) return;

        const sessionIndex = state.sessions.findIndex(session => session.id === action.payload.sessionId);
        if (sessionIndex === -1) return;

        state.sessions.splice(sessionIndex, 1);

        if (state.activeSessionId === action.payload.sessionId) {
            state.activeSessionId = state.sessions[0]?.id || null;
        }
    },
    setActiveSession: (state, action: PayloadAction<{ sessionId: string }>) => {
        const session = state.sessions.find(s => s.id === action.payload.sessionId);
        if (session) {
            state.activeSessionId = action.payload.sessionId;
        }
    },
    updateSessionName: (state, action: PayloadAction<{ sessionId: string; name: string }>) => {
        const session = state.sessions.find(s => s.id === action.payload.sessionId);
        if (session) {
            session.name = action.payload.name;
        }
    },
    clearAllChatSessions: (state) => {
        state.sessions = [];
        state.activeSessionId = null;
    },
  },
});

export const HoudiniActions = houdiniSlice.actions;
export const houdiniReducers = houdiniSlice.reducer;