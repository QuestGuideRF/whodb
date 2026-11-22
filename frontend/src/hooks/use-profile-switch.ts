import { useMutation } from '@apollo/client/react';
import { useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from '@clidey/ux';
import { LoginSourceDocument, LoginWithSourceProfileDocument } from '@graphql';
import { useAppDispatch } from '@/store/hooks';
import { AuthActions, LocalLoginProfile } from '@/store/auth';
import { DatabaseActions } from '@/store/database';
import { updateProfileLastAccessed } from '@/components/profile-info-tooltip';
import { InternalRoutes } from '@/config/routes';
import { clearGraphqlStore } from '@/config/graphql-client';

interface UseProfileSwitchOptions {
    onSuccess?: () => void;
    onError?: (error: string) => void;
    errorMessage?: string;
}

export const useProfileSwitch = (options?: UseProfileSwitchOptions) => {
    const dispatch = useAppDispatch();
    const navigate = useNavigate();
    const [login, { loading: loginLoading }] = useMutation(LoginSourceDocument);
    const [loginWithSourceProfile, { loading: loginWithSourceProfileLoading }] = useMutation(LoginWithSourceProfileDocument);

    const loading = loginLoading || loginWithSourceProfileLoading;

    const switchProfile = useCallback(async (profile: LocalLoginProfile, database?: string) => {
        const targetDatabase = database ?? profile.Database;

        dispatch(DatabaseActions.setSchema(""));

        try {
            const switchSucceeded = profile.Saved || profile.IsEnvironmentDefined
                ? (await loginWithSourceProfile({
                    variables: {
                        profile: {
                            Id: profile.Id,
                            Values: targetDatabase ? [{ Key: "Database", Value: targetDatabase }] : [],
                        }
                    },
                })).data?.LoginWithSourceProfile.Status
                : (await login({
                    variables: {
                        credentials: {
                            Id: profile.Id,
                            SourceType: profile.SourceType,
                            Values: profile.Database === targetDatabase
                                ? profile.Values
                                : profile.Values.map(value =>
                                    value.Key === "Database" ? { ...value, Value: targetDatabase } : value
                                  ).concat(
                                    profile.Values.some(value => value.Key === "Database")
                                      ? []
                                      : [{ Key: "Database", Value: targetDatabase }]
                                  ),
                            AccessToken: profile.AccessToken,
                        }
                    },
                })).data?.LoginSource.Status;

            if (!switchSucceeded) {
                const errorMsg = options?.errorMessage ?? 'Failed to switch profile';
                toast.error(errorMsg);
                options?.onError?.(errorMsg);
                return;
            }

            updateProfileLastAccessed(profile.Id);
            if (database) {
                dispatch(AuthActions.setLoginProfileDatabase({ id: profile.Id, database }));
            }
            await clearGraphqlStore();
            dispatch(DatabaseActions.setSchema(""));
            dispatch(AuthActions.switch({ id: profile.Id }));
            navigate(InternalRoutes.Dashboard.StorageUnit.path, {
                state: {},
            });
            options?.onSuccess?.();
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : String(error);
            const errorMsg = `${options?.errorMessage ?? 'Failed to switch profile'}: ${errorMessage}`;
            toast.error(errorMsg);
            options?.onError?.(errorMessage);
        }
    }, [dispatch, login, loginWithSourceProfile, navigate, options]);

    return {
        switchProfile,
        loading,
    };
};