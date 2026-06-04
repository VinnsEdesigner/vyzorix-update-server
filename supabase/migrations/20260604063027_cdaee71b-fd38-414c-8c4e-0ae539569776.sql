
REVOKE EXECUTE ON FUNCTION public.has_admin_access(text) FROM PUBLIC, anon, authenticated;
GRANT EXECUTE ON FUNCTION public.has_admin_access(text) TO service_role;
