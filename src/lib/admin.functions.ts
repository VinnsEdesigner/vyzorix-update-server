import { createServerFn } from "@tanstack/react-start";
import { requireSupabaseAuth } from "@/integrations/supabase/auth-middleware";

/**
 * Self-bootstrapping admin allowlist:
 * - If the allowlist is empty, the first authenticated user becomes the
 *   sole admin and signups are locked from then on.
 * - Otherwise, only allowlisted emails are granted access. Anyone else
 *   gets `{ allowed: false }` and the client signs them out.
 */
export const ensureAdminAccess = createServerFn({ method: "POST" })
  .middleware([requireSupabaseAuth])
  .handler(async ({ context }) => {
    const { supabaseAdmin } = await import("@/integrations/supabase/client.server");
    const email = (context.claims?.email as string | undefined)?.toLowerCase();
    if (!email) return { allowed: false, reason: "no_email" as const };

    const { count, error: countErr } = await supabaseAdmin
      .from("app_admins")
      .select("email", { count: "exact", head: true });
    if (countErr) throw new Error(countErr.message);

    if ((count ?? 0) === 0) {
      const { error: insertErr } = await supabaseAdmin
        .from("app_admins")
        .insert({ email, user_id: context.userId });
      if (insertErr) throw new Error(insertErr.message);
      // Lock the door behind the first admin.
      try {
        await fetch(`${process.env.SUPABASE_URL}/auth/v1/admin/config`, {
          method: "PATCH",
          headers: {
            apikey: process.env.SUPABASE_SERVICE_ROLE_KEY!,
            Authorization: `Bearer ${process.env.SUPABASE_SERVICE_ROLE_KEY}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ DISABLE_SIGNUP: true }),
        });
      } catch {
        // Best effort; the allowlist still enforces access at the app boundary.
      }
      return { allowed: true, bootstrapped: true, email };
    }

    const { data, error } = await supabaseAdmin
      .from("app_admins")
      .select("email")
      .eq("email", email)
      .maybeSingle();
    if (error) throw new Error(error.message);

    return { allowed: !!data, bootstrapped: false, email };
  });