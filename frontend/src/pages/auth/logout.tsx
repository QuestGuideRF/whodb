import { useMutation } from "@apollo/client/react";
import { FC, useEffect } from "react";
import { useDispatch } from "react-redux";
import { Container } from "../../components/page";
import { LogoutDocument, LogoutMutation, LogoutMutationVariables } from '@graphql';
import { AuthActions } from "../../store/auth";
import { Loading } from "../../components/loading";
import { toast } from "@clidey/ux";
import { useTranslation } from '@/hooks/use-translation';
import { clearGraphqlStore } from "@/config/graphql-client";

export const LogoutPage: FC = () => {
  const { t } = useTranslation('pages/logout');
  const [logout, ] = useMutation<LogoutMutation, LogoutMutationVariables>(LogoutDocument);
  const dispatch = useDispatch();

  useEffect(() => {
    void (async () => {
      try {
        await logout();
        await clearGraphqlStore();
        dispatch(AuthActions.logout());
        toast.success(t('success'));
      } catch {
        toast.error(t('error'));
      }
    })();
  }, [dispatch, logout, t]);

  return <Container>
      <div className="flex flex-col justify-center items-center gap-lg w-full">
          <div>
              <Loading hideText={true} />
          </div>
          <div className="text-neutral-800 dark:text-neutral-300">
              {t('loggingOut')}
          </div>
      </div>
  </Container>
}