interface LogConsoleProps {
  height?: string;
  className?: string;
}

export function LogConsole({ height = "h-64", className = "" }: LogConsoleProps) {
  return (
    <div
      className={`${height} ${className} overflow-auto p-4 font-mono text-xs bg-black text-green-400`}
    >
      <div className="text-muted-foreground">[console ready]</div>
    </div>
  );
}
