import {createSlice, PayloadAction} from '@reduxjs/toolkit';
import {SortCondition, WhereCondition} from '@graphql';

type ExploreConditionEntry = {
    whereCondition?: WhereCondition;
    sortConditions: SortCondition[];
};

type IExploreConditionsState = {
    conditions: Record<string, ExploreConditionEntry>;
};

const initialState: IExploreConditionsState = {
    conditions: {},
};

export const exploreConditionsSlice = createSlice({
    name: 'exploreConditions',
    initialState,
    reducers: {
        setExploreConditions: (state, action: PayloadAction<{ key: string } & ExploreConditionEntry>) => {
            const { key, whereCondition, sortConditions } = action.payload;
            state.conditions[key] = { whereCondition, sortConditions };
        },
    },
});

export const ExploreConditionsActions = exploreConditionsSlice.actions;
export const exploreConditionsReducers = exploreConditionsSlice.reducer;