import { PayloadAction, createSlice } from '@reduxjs/toolkit';

export interface ITourState {
    isActive: boolean;
    tourId: string | null;
    shouldStartOnLoad: boolean;
}

const initialState: ITourState = {
    isActive: false,
    tourId: null,
    shouldStartOnLoad: false,
};

export const tourSlice = createSlice({
    name: 'tour',
    initialState,
    reducers: {
        startTour: (state, action: PayloadAction<string>) => {
            state.isActive = true;
            state.tourId = action.payload;
            state.shouldStartOnLoad = false;
        },
        stopTour: (state) => {
            state.isActive = false;
            state.tourId = null;
            state.shouldStartOnLoad = false;
        },
        scheduleTourOnLoad: (state, action: PayloadAction<string>) => {
            state.shouldStartOnLoad = true;
            state.tourId = action.payload;
        },
    },
});

export const TourActions = tourSlice.actions;
export const tourReducers = tourSlice.reducer;
export default tourSlice.reducer;