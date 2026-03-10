type ShellResult = {
  stdout: string;
  stderr: string;
  exitCode: number;
};

export function run(command: string, args: string[]): ShellResult {
  const proc = Bun.spawnSync([command, ...args], {
    stdout: "pipe",
    stderr: "pipe",
    stdin: "ignore",
  });

  return {
    stdout: new TextDecoder().decode(proc.stdout).trim(),
    stderr: new TextDecoder().decode(proc.stderr).trim(),
    exitCode: proc.exitCode,
  };
}

export function mustRun(command: string, args: string[]): string {
  const result = run(command, args);
  if (result.exitCode !== 0) {
    const detail = result.stderr || result.stdout || `exit ${result.exitCode}`;
    throw new Error(`${command} ${args.join(" ")} failed: ${detail}`);
  }
  return result.stdout;
}
