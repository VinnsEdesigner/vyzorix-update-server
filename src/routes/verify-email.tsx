import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { CheckCircle, Loader2, AlertCircle } from "lucide-react";
import { useState, useEffect, type ReactElement } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Toaster } from "@/components/ui/sonner";
import { verifyEmail } from "@/lib/vyzorix-auth";
import { useVyzorixConfig } from "@/lib/vyzorix-config";

export const Route = createFileRoute("/verify-email")({
  ssr: false,
  head: () => ({ meta: [{ title: "Verify Email — Vyzorix" }] }),
  component: VerifyEmailPage,
});

type Status = "loading" | "success" | "error";

// eslint-disable-next-line func-style
function VerifyEmailPage(): ReactElement {
  const navigate = useNavigate();
  const { serverUrl } = useVyzorixConfig();
  const [status, setStatus] = useState<Status>("loading");
  const [message, setMessage] = useState("");

  useEffect(() => {
    // eslint-disable-next-line @typescript-eslint/explicit-function-return-type
    const verify = async () => {
      const params = new URLSearchParams(window.location.search);
      const token = params.get("token");

      if (!token) {
        setStatus("error");
        setMessage("Invalid verification link. Please request a new verification email.");
        return;
      }

      try {
        await verifyEmail(serverUrl, token);
        setStatus("success");
        setMessage("Your email has been verified successfully!");
        toast.success("Email verified!");
      } catch (err) {
        setStatus("error");
        setMessage(err instanceof Error ? err.message : "Failed to verify email");
        toast.error("Verification failed");
      }
    };

    verify();
  }, [serverUrl]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/30 px-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
            {status === "loading" && <Loader2 className="h-6 w-6 animate-spin text-primary" />}
            {status === "success" && <CheckCircle className="h-6 w-6 text-green-500" />}
            {status === "error" && <AlertCircle className="h-6 w-6 text-destructive" />}
          </div>
          <CardTitle>
            {status === "loading" && "Verifying..."}
            {status === "success" && "Email Verified!"}
            {status === "error" && "Verification Failed"}
          </CardTitle>
          <CardDescription>{message}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {status === "loading" && (
            <p className="text-center text-sm text-muted-foreground">
              Please wait while we verify your email address...
            </p>
          )}
          {status === "success" && (
            <Button className="w-full" onClick={() => navigate({ to: "/dashboard" })}>
              Go to Dashboard
            </Button>
          )}
          {status === "error" && (
            <div className="space-y-3">
              <Button
                variant="outline"
                className="w-full"
                onClick={() => navigate({ to: "/login" })}
              >
                Return to Login
              </Button>
              <Button
                variant="ghost"
                className="w-full"
                onClick={() => navigate({ to: "/forgot-password" })}
              >
                Request New Verification
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
      <Toaster />
    </div>
  );
}
