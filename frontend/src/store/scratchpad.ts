import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { WhereCondition } from '@graphql';
import { v4 as uuidv4 } from 'uuid';

export type ScratchpadPage = {
  id: string;
  name: string;
  cellIds: string[];
  conditions?: WhereCondition;
}

export type ScratchpadCell = {
  id: string;
  code: string;
  mode: string;
  history: Array<{
    id: string;
    item: string;
    status: boolean;
    date: Date;
  }>;
}

export type IScratchpadState = {
  pages: ScratchpadPage[];
  cells: Record<string, ScratchpadCell>;
  activePageId: string | null;
}

const initialState: IScratchpadState = {
  pages: [],
  cells: {},
  activePageId: null,
}

export const scratchpadSlice = createSlice({
  name: 'scratchpad',
  initialState,
  reducers: {
    initializeScratchpad: (state) => {
      if (state.pages.length === 0) {
        const newId = uuidv4();
        const cellId = uuidv4();
        const firstPage: ScratchpadPage = { id: newId, name: "Page 1", cellIds: [cellId] };
        const firstCell: ScratchpadCell = {
          id: cellId,
          code: "",
          mode: "Query",
          history: []
        };
        state.pages = [firstPage];
        state.cells[cellId] = firstCell;
        state.activePageId = newId;
      }
    },
    addPage: (state, action: PayloadAction<{ name?: string; initialQuery?: string }>) => {
      const newId = uuidv4();
      const cellId = uuidv4();
      const newPage: ScratchpadPage = {
        id: newId,
        name: action.payload.name || `Page ${state.pages.length + 1}`,
        cellIds: [cellId]
      };
      const newCell: ScratchpadCell = {
        id: cellId,
        code: action.payload.initialQuery || "",
        mode: "Query",
        history: []
      };
      state.pages.push(newPage);
      state.cells[cellId] = newCell;
      state.activePageId = newId;
    },
    deletePage: (state, action: PayloadAction<{ pageId: string }>) => {
      if (state.pages.length <= 1) return;

      const pageIndex = state.pages.findIndex(page => page.id === action.payload.pageId);
      if (pageIndex === -1) return;

      const page = state.pages[pageIndex];
      page.cellIds.forEach(cellId => {
        delete state.cells[cellId];
      });

      state.pages.splice(pageIndex, 1);

      if (state.activePageId === action.payload.pageId) {
        state.activePageId = state.pages[0]?.id || null;
      }
    },
    setActivePage: (state, action: PayloadAction<{ pageId: string }>) => {
      const page = state.pages.find(p => p.id === action.payload.pageId);
      if (page) {
        state.activePageId = action.payload.pageId;
      }
    },
    updatePageName: (state, action: PayloadAction<{ pageId: string; name: string }>) => {
      const page = state.pages.find(p => p.id === action.payload.pageId);
      if (page) {
        page.name = action.payload.name;
      }
    },
    addCell: (state, action: PayloadAction<{ pageId: string; afterCellId?: string; initialQuery?: string }>) => {
      const newCellId = uuidv4();
      const newCell: ScratchpadCell = {
        id: newCellId,
        code: action.payload.initialQuery || "",
        mode: "Query",
        history: []
      };

      state.cells[newCellId] = newCell;

      const page = state.pages.find(p => p.id === action.payload.pageId);
      if (page) {
        if (action.payload.afterCellId) {
          const index = page.cellIds.indexOf(action.payload.afterCellId);
          if (index !== -1) {
            page.cellIds.splice(index + 1, 0, newCellId);
          } else {
            page.cellIds.push(newCellId);
          }
        } else {
          page.cellIds.push(newCellId);
        }
      }
    },
    deleteCell: (state, action: PayloadAction<{ pageId: string; cellId: string }>) => {
      const page = state.pages.find(p => p.id === action.payload.pageId);
      if (page && page.cellIds.length > 1) {
        page.cellIds = page.cellIds.filter(id => id !== action.payload.cellId);
        delete state.cells[action.payload.cellId];
      }
    },
    updateCellCode: (state, action: PayloadAction<{ cellId: string; code: string }>) => {
      const cell = state.cells[action.payload.cellId];
      if (cell) {
        cell.code = action.payload.code;
      }
    },
    updateCellMode: (state, action: PayloadAction<{ cellId: string; mode: string }>) => {
      const cell = state.cells[action.payload.cellId];
      if (cell) {
        cell.mode = action.payload.mode;
      }
    },
    addCellHistory: (state, action: PayloadAction<{ cellId: string; item: string; status: boolean }>) => {
      const cell = state.cells[action.payload.cellId];
      if (cell) {
        cell.history.unshift({
          id: uuidv4(),
          item: action.payload.item,
          status: action.payload.status,
          date: new Date()
        });
      }
    },
    clearCellHistory: (state, action: PayloadAction<{ cellId: string }>) => {
      const cell = state.cells[action.payload.cellId];
      if (cell) {
        cell.history = [];
      }
    },
    clearAllScratchpad: (state) => {
      state.pages = [];
      state.cells = {};
      state.activePageId = null;
    },
    ensurePagesHaveCells: (state) => {

      state.pages.forEach(page => {

        if (!page.cellIds || page.cellIds.length === 0) {
          const cellId = uuidv4();
          const newCell: ScratchpadCell = {
            id: cellId,
            code: "",
            mode: "Query",
            history: []
          };

          page.cellIds = [cellId];
          state.cells[cellId] = newCell;
        }
      });

      if (state.pages.length === 0) {
        const newId = uuidv4();
        const cellId = uuidv4();
        const firstPage: ScratchpadPage = { id: newId, name: "Page 1", cellIds: [cellId] };
        const firstCell: ScratchpadCell = {
          id: cellId,
          code: "",
          mode: "Query",
          history: []
        };
        state.pages = [firstPage];
        state.cells[cellId] = firstCell;
        state.activePageId = newId;
      }

      if (!state.activePageId && state.pages.length > 0) {
        state.activePageId = state.pages[0].id;
      }
    },
    addConditionToPage: (state, action: PayloadAction<{ pageId: string; condition: WhereCondition }>) => {
      const page = state.pages.find(p => p.id === action.payload.pageId);
      if (page) {
        page.conditions = action.payload.condition;
      }
    },
    removeConditionFromPage: (state, action: PayloadAction<{ pageId: string }>) => {
      const page = state.pages.find(p => p.id === action.payload.pageId);
      if (page) {
        delete page.conditions;
      }
    },
    addCellToPageAndActivate: (state, action: PayloadAction<{ pageId: string; initialQuery?: string }>) => {
      const newCellId = uuidv4();
      const newCell: ScratchpadCell = {
        id: newCellId,
        code: action.payload.initialQuery || "",
        mode: "Query",
        history: []
      };

      state.cells[newCellId] = newCell;

      const page = state.pages.find(p => p.id === action.payload.pageId);
      if (page) {
        page.cellIds.push(newCellId);

        state.activePageId = action.payload.pageId;
      }
    }
  },
});

export const ScratchpadActions = scratchpadSlice.actions;
export const scratchpadReducers = scratchpadSlice.reducer;