
CREATE TABLE public.app_admins (
  email text PRIMARY KEY,
  user_id uuid REFERENCES auth.users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

GRANT SELECT ON public.app_admins TO authenticated;
GRANT ALL ON public.app_admins TO service_role;

ALTER TABLE public.app_admins ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Admins can view allowlist"
  ON public.app_admins FOR SELECT
  TO authenticated
  USING (email = (auth.jwt() ->> 'email'));

CREATE OR REPLACE FUNCTION public.has_admin_access(_email text)
RETURNS boolean
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = public
AS $$
  SELECT EXISTS (SELECT 1 FROM public.app_admins WHERE email = _email);
$$;
