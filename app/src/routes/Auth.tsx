import { tokenField } from "@/conf/bootstrap.ts";
import { useEffect, useReducer } from "react";
import Loader from "@/components/Loader.tsx";
import "@/assets/pages/auth.less";
import { validateToken } from "@/store/auth.ts";
import { useDispatch, useSelector } from "react-redux";
import router from "@/router.tsx";
import { useTranslation } from "react-i18next";
import { getQueryParam } from "@/utils/path.ts";
import { setMemory } from "@/utils/memory.ts";
import { appLogo, appName, useDeeptrain } from "@/conf/env.ts";
import { Card, CardContent } from "@/components/ui/card.tsx";
import { goAuth } from "@/utils/app.ts";
import { Label } from "@/components/ui/label.tsx";
import { Input } from "@/components/ui/input.tsx";
import Require, { LengthRangeRequired } from "@/components/Require.tsx";
import { Button } from "@/components/ui/button.tsx";
import { formReducer, isTextInRange } from "@/utils/form.ts";
import { doLogin, LoginForm } from "@/api/auth.ts";
import { getErrorMessage, isEnter } from "@/utils/base.ts";
import { ScrollArea } from "@/components/ui/scroll-area.tsx";
import { toast } from "sonner";
import {
  infoLinuxDoOAuthSelector,
  infoGitHubOAuthSelector,
} from "@/store/info.ts";
import { Github } from "lucide-react";

function DeepAuth() {
  const { t } = useTranslation();
  const dispatch = useDispatch();
  const token = getQueryParam("token").trim();

  useEffect(() => {
    if (!token.length) {
      toast.warning(t("invalid-token"), {
        description: t("invalid-token-prompt"),
        action: {
          label: t("try-again"),
          onClick: goAuth,
        },
      });

      setTimeout(goAuth, 2500);
      return;
    }

    setMemory(tokenField, token);

    doLogin({ token })
      .then((data) => {
        if (!data.status) {
          toast.error(t("login-failed"), {
            description: t("login-failed-prompt", { reason: data.error }),
            action: {
              label: t("try-again"),
              onClick: goAuth,
            },
          });
        } else
          validateToken(dispatch, data.token, async () => {
            toast.success(t("login-success"), {
              description: t("login-success-prompt"),
            });

            await router.navigate("/");
          });
      })
      .catch((err) => {
        console.debug(err);

        toast.error(t("server-error"), {
          description: `${t("server-error-prompt")}\n${err.message}`,
          action: {
            label: t("try-again"),
            onClick: goAuth,
          },
        });
      });
  }, []);

  return (
    <div className={`auth`}>
      <Loader prompt={t("login")} />
    </div>
  );
}

function Login() {
  const { t } = useTranslation();
  const globalDispatch = useDispatch();
  const [form, dispatch] = useReducer(formReducer<LoginForm>(), {
    username: sessionStorage.getItem("username") || "",
    password: sessionStorage.getItem("password") || "",
  });

  const onSubmit = async () => {
    if (
      !isTextInRange(form.username, 1, 255) ||
      !isTextInRange(form.password, 6, 36)
    )
      return;

    try {
      const resp = await doLogin(form);
      if (!resp.status) {
        toast.warning(t("login-failed"), {
          description: t("login-failed-prompt", { reason: resp.error }),
        });
        return;
      }

      toast.success(t("login-success"), {
        description: t("login-success-prompt"),
      });

      if (
        form.username.trim() === "root" &&
        form.password.trim() === "neoai123456"
      ) {
        toast.warning(t("admin.default-password"), {
          description: t("admin.default-password-prompt"),
          duration: 15000,
        });
      }

      validateToken(globalDispatch, resp.token);
      await router.navigate("/");
    } catch (err) {
      console.debug(err);
      toast.error(t("server-error"), {
        description: `${t("server-error-prompt")}\n${getErrorMessage(err)}`,
      });
    }
  };

  useEffect(() => {
    // listen to enter key and auto submit
    const listener = async (e: KeyboardEvent) => {
      if (isEnter(e)) await onSubmit();
    };

    document.addEventListener("keydown", listener);
    return () => document.removeEventListener("keydown", listener);
  }, []);

  return (
    <ScrollArea className={`w-full h-full grid place-items-center`}>
      <div className={`auth-container`}>
        <img className={`logo`} src={appLogo} alt="" />
        <div className={`title`}>
          {t("login")} {appName}
        </div>
        <Card className={`auth-card`}>
          <CardContent className={`pb-0`}>
            <div className={`auth-wrapper`}>
              <Label>
                <Require />
                {t("auth.username-or-email")}
                <LengthRangeRequired
                  content={form.username}
                  min={1}
                  max={255}
                  hideOnEmpty={true}
                />
              </Label>
              <Input
                placeholder={t("auth.username-or-email-placeholder")}
                value={form.username}
                onChange={(e) =>
                  dispatch({ type: "update:username", payload: e.target.value })
                }
              />

              <Label>
                <Require />
                {t("auth.password")}
                <LengthRangeRequired
                  content={form.password}
                  min={6}
                  max={36}
                  hideOnEmpty={true}
                />
              </Label>
              <Input
                placeholder={t("auth.password-placeholder")}
                value={form.password}
                type={"password"}
                onChange={(e) =>
                  dispatch({ type: "update:password", payload: e.target.value })
                }
              />

              <Button
                tapScale={0.975}
                classNameWrapper={`mt-2`}
                onClick={onSubmit}
                className={`w-full`}
                loading={true}
              >
                {t("login")}
              </Button>

              <OAuthButtons />
            </div>
          </CardContent>
        </Card>
        <div className={`auth-card addition-wrapper`}>
          <div className={`row`}>
            {t("auth.no-account")}
            <a className={`link`} onClick={() => router.navigate("/register")}>
              {t("auth.register")}
            </a>
          </div>
          <div className={`row`}>
            {t("auth.forgot-password")}
            <a className={`link`} onClick={() => router.navigate("/forgot")}>
              {t("auth.reset-password")}
            </a>
          </div>
        </div>
      </div>
    </ScrollArea>
  );
}

function Auth() {
  return useDeeptrain ? <DeepAuth /> : <Login />;
}

// OAuthButtons — renders LinuxDO / GitHub login buttons if those OAuth
// providers are configured on the backend. The /info endpoint exposes
// `linuxdo_oauth_enabled` and `github_oauth_enabled` flags.
//
// Clicking a button navigates to /api/oauth/<provider>/login which
// redirects to the provider's authorize URL. The OAuth callback then
// redirects back to the frontend with #access_token=<jwt> which the
// bootstrap reads and stores.
function OAuthButtons() {
  const linuxDoEnabled = useSelector(infoLinuxDoOAuthSelector);
  const githubEnabled = useSelector(infoGitHubOAuthSelector);
  const apiBase =
    (axios.defaults.baseURL as string | undefined) || "/api";

  if (!linuxDoEnabled && !githubEnabled) return null;

  return (
    <div className="mt-3 space-y-2">
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <div className="h-px bg-border flex-1" />
        <span>or continue with</span>
        <div className="h-px bg-border flex-1" />
      </div>
      <div className="grid grid-cols-2 gap-2">
        {linuxDoEnabled && (
          <Button
            variant="outline"
            className="w-full"
            onClick={() => {
              window.location.href = `${apiBase}/oauth/linuxdo/login`;
            }}
          >
            <span className="mr-1.5 font-bold text-[#f5a623]">L</span>
            LinuxDO
          </Button>
        )}
        {githubEnabled && (
          <Button
            variant="outline"
            className="w-full"
            onClick={() => {
              window.location.href = `${apiBase}/oauth/github/login`;
            }}
          >
            <Github className="h-4 w-4 mr-1.5" />
            GitHub
          </Button>
        )}
      </div>
    </div>
  );
}

// silence unused-import warnings on axios if not already imported
import axios from "axios";

export default Auth;
