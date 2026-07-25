

import { commands } from "./commands-endpoints";
import type { Command, CommandStatus } from "@/domain/commands";
import { isCommandTerminal } from "@/domain/commands";

export interface CommandPollingOptions {
  pollInterval?: number;
  deliveryTimeout?: number;
  completionTimeout?: number;
}

export interface CommandPollingCallbacks {
  onPending?: (command: Command) => void;
  onDelivered?: (command: Command) => void;
  onCompleted?: (command: Command) => void;
  onFailed?: (command: Command, error?: string) => void;
  onCancelled?: (command: Command) => void;
  onTimeout?: (command: Command, stage: 'delivery' | 'completion') => void;
  onError?: (error: Error) => void;
  onStateChange?: (previousStatus: CommandStatus, currentStatus: CommandStatus, command: Command) => void;
}

export interface CommandPoller {
  stop(): void;
  getLatestStatus(): CommandStatus | null;
  isRunning(): boolean;
}

export function pollCommandStatus(
  dispatchId: string,
  options: CommandPollingOptions = {},
  callbacks: CommandPollingCallbacks = {}
): CommandPoller {
  const {
    pollInterval = 2000,
    deliveryTimeout = 30000,
    completionTimeout = 120000,
  } = options;

  let intervalId: ReturnType<typeof setInterval> | null = null;
  let timeoutId: ReturnType<typeof setTimeout> | null = null;
  let isActive = true;
  let lastStatus: CommandStatus | null = null;

  function stop() {
    isActive = false;
    if (intervalId) clearInterval(intervalId);
    if (timeoutId) clearTimeout(timeoutId);
  }

  function fetchAndProcess() {
    if (!isActive) return;

    commands.getByDispatchId(dispatchId)
      .then((command) => {
        if (!command) {
          callbacks.onError?.(new Error(`Command ${dispatchId} not found`));
          stop();
          return;
        }

        const prevStatus = lastStatus;
        if (prevStatus && prevStatus !== command.status) {
          callbacks.onStateChange?.(prevStatus, command.status, command);
        }
        lastStatus = command.status;

        switch (command.status) {
          case "pending":
            callbacks.onPending?.(command);
            break;
          case "delivered":
            callbacks.onDelivered?.(command);
            if (timeoutId) clearTimeout(timeoutId);
            timeoutId = setTimeout(() => {
              if (isActive) {
                commands.getByDispatchId(dispatchId).then((cmd) => {
                  if (cmd && !isCommandTerminal(cmd.status)) {
                    callbacks.onTimeout?.(cmd, 'completion');
                    stop();
                  }
                });
              }
            }, completionTimeout);
            break;
          case "completed":
            callbacks.onCompleted?.(command);
            stop();
            break;
          case "failed":
            callbacks.onFailed?.(command, command.failureReason);
            stop();
            break;
          case "cancelled":
            callbacks.onCancelled?.(command);
            stop();
            break;
        }

        if (isCommandTerminal(command.status)) {
          stop();
        }
      })
      .catch((error) => {
        callbacks.onError?.(error);
      });
  }

  fetchAndProcess();
  intervalId = setInterval(fetchAndProcess, pollInterval);
  timeoutId = setTimeout(() => {
    if (isActive) {
      commands.getByDispatchId(dispatchId).then((cmd) => {
        if (cmd && !isCommandTerminal(cmd.status)) {
          callbacks.onTimeout?.(cmd, 'delivery');
          stop();
        }
      });
    }
  }, deliveryTimeout);

  return { stop, getLatestStatus: () => lastStatus, isRunning: () => isActive };
}

export async function waitForCommandCompletion(
  dispatchId: string,
  options: CommandPollingOptions = {}
): Promise<Command> {
  return new Promise((resolve, reject) => {
    pollCommandStatus(dispatchId, options, {
      onCompleted: (cmd) => resolve(cmd),
      onFailed: (cmd, error) => reject(new Error(error ?? "Command failed")),
      onCancelled: () => reject(new Error("Command was cancelled")),
      onTimeout: (_, stage) => reject(new Error(`Command ${stage} timed out`)),
      onError: (err) => reject(err),
    });
  });
}
