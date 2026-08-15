import { createFileRoute } from '@tanstack/react-router';

function IndexRoute() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="text-center">
        <h1 className="text-2xl font-semibold">VyzoriX</h1>
        <p className="text-muted-foreground">Runtime ready</p>
      </div>
    </div>
  );
}

export const Route = createFileRoute('/')({
  component: IndexRoute,
});
