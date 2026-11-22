import {skipToken, useQuery} from "@apollo/client/react";
import { FC } from "react";
import { GetAiModelsDocument } from '@graphql';
import { Loading } from "../../components/loading";
import { Navigate } from "react-router-dom";
import { InternalRoutes } from "../../config/routes";
import { useSourceContract } from "../../hooks/useSourceContract";
import { InternalPage } from "../../components/page";
import { useAppSelector } from "../../store/hooks";
import { availableInternalModelTypes } from "../../store/ai-models";
import { hasComponent } from "../../config/component-registry";

export const NavigateToDefault: FC = () => {
    const current = useAppSelector(state => state.auth.current);
    const { supportsChat } = useSourceContract(current?.Type);

    if (hasComponent('sql-agent')) {
        return <Navigate to={InternalRoutes.Chat.path} />
    }

    const defaultModelType = availableInternalModelTypes[0];
    const aiModelsQueryOptions = current && supportsChat && defaultModelType
        ? {
            variables: {
                modelType: defaultModelType,
            }
        }
        : skipToken;
    const { data, error } = useQuery(GetAiModelsDocument, aiModelsQueryOptions);

    if (!supportsChat ||  error != null) {
        return <Navigate to={InternalRoutes.Dashboard.StorageUnit.path} />
    }

    if (data?.AIModel != null) {
        if (data.AIModel.length > 0) {
            return <Navigate to={InternalRoutes.Chat.path} />
        }
        return <Navigate to={InternalRoutes.Dashboard.StorageUnit.path} />
    }

    return <InternalPage>
        <Loading />
    </InternalPage>
  }