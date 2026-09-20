import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import {
  azerothMeAccountCreateMutation,
  azerothMeAccountGetOptions,
  azerothMeAccountGetQueryKey,
} from "@/api";
import { isApiError } from "@/api/errors";
import { TextField } from "@/components/common/form-controls";
import { LoadingState } from "@/components/common/loading-state";
import { PageHeader } from "@/components/common/page-header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Form } from "@/components/ui/form";

const schema = z.object({
  username: z.string().min(1, "Required").max(32),
  password: z.string().min(1, "Required").max(64),
});

type FormValues = z.infer<typeof schema>;

/** OnboardingPage lets a user create and link a game account. */
export function OnboardingPage() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const query = useQuery(azerothMeAccountGetOptions());
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { username: "", password: "" },
  });
  const mutation = useMutation(azerothMeAccountCreateMutation());

  if (query.isPending) {
    return <LoadingState label="Loading account…" />;
  }

  const submit = form.handleSubmit(async (values) => {
    try {
      await mutation.mutateAsync({ body: values });
      toast.success("Account created and linked");
      await queryClient.invalidateQueries({ queryKey: azerothMeAccountGetQueryKey() });
      await navigate({ to: "/characters" });
    } catch (error) {
      toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));
    }
  });

  return (
    <div className="mx-auto max-w-3xl p-8">
      <PageHeader
        title="Game account"
        description="Link a game account to use characters, mail and the store."
      />

      {query.data?.linked ? (
        <Card>
          <CardHeader>
            <CardTitle>Account linked</CardTitle>
            <CardDescription>
              Linked to {query.data.account_username ?? "your account"}.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button onClick={() => void navigate({ to: "/characters" })}>My characters</Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Create a new account</CardTitle>
              <CardDescription>Choose your own username and password.</CardDescription>
            </CardHeader>
            <CardContent>
              <Form {...form}>
                <form onSubmit={submit} className="space-y-4">
                  <TextField control={form.control} name="username" label="Username" />
                  <TextField
                    control={form.control}
                    name="password"
                    label="Password"
                    type="password"
                  />
                  <Button type="submit" disabled={mutation.isPending}>
                    Create and link
                  </Button>
                </form>
              </Form>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>I already have an account</CardTitle>
              <CardDescription>
                Claim an existing account with an in-game code. Coming soon.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Button variant="outline" disabled>
                Claim account
              </Button>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
