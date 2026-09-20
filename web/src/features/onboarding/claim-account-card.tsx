import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";

import {
  azerothMeAccountClaimStartMutation,
  azerothMeAccountClaimVerifyMutation,
  azerothMeAccountGetQueryKey,
} from "@/api";
import { isApiError } from "@/api/errors";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

/** ClaimAccountCard claims an existing account with an in-game code. */
export function ClaimAccountCard() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [step, setStep] = useState<"start" | "verify">("start");
  const [accountUsername, setAccountUsername] = useState("");
  const [character, setCharacter] = useState("");
  const [code, setCode] = useState("");

  const start = useMutation(azerothMeAccountClaimStartMutation());
  const verify = useMutation(azerothMeAccountClaimVerifyMutation());

  const report = (error: unknown) =>
    toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));

  const onStart = async () => {
    try {
      await start.mutateAsync({ body: { account_username: accountUsername, character } });
      toast.success("Code sent in game");
      setStep("verify");
    } catch (error) {
      report(error);
    }
  };

  const onVerify = async () => {
    try {
      await verify.mutateAsync({ body: { account_username: accountUsername, code } });
      toast.success("Account linked");
      await queryClient.invalidateQueries({ queryKey: azerothMeAccountGetQueryKey() });
      await navigate({ to: "/characters" });
    } catch (error) {
      report(error);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>I already have an account</CardTitle>
        <CardDescription>
          {step === "start"
            ? "We mail a one-time code to one of your characters."
            : "Enter the code from the in-game mail."}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {step === "start" ? (
          <>
            <Input
              placeholder="Account username"
              aria-label="Account username"
              value={accountUsername}
              onChange={(event) => setAccountUsername(event.target.value)}
            />
            <Input
              placeholder="Character name"
              aria-label="Character name"
              value={character}
              onChange={(event) => setCharacter(event.target.value)}
            />
            <Button
              type="button"
              onClick={() => void onStart()}
              disabled={start.isPending || !accountUsername || !character}
            >
              Send code
            </Button>
          </>
        ) : (
          <>
            <Input
              placeholder="6-digit code"
              aria-label="Claim code"
              value={code}
              onChange={(event) => setCode(event.target.value)}
            />
            <div className="flex gap-2">
              <Button
                type="button"
                onClick={() => void onVerify()}
                disabled={verify.isPending || !code}
              >
                Verify and link
              </Button>
              <Button type="button" variant="outline" onClick={() => setStep("start")}>
                Back
              </Button>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
}
