'use client';

import * as Select from '@radix-ui/react-select';
import { Check, ChevronDown } from 'lucide-react';

export type HeadlessSelectOption = {
  value: string;
  label: string;
  description?: string;
};

type HeadlessSelectProps = {
  value: string;
  onValueChange: (value: string) => void;
  options: HeadlessSelectOption[];
  ariaLabel: string;
  placeholder?: string;
  disabled?: boolean;
  triggerClassName?: string;
};

export function HeadlessSelect({
  value,
  onValueChange,
  options,
  ariaLabel,
  placeholder,
  disabled = false,
  triggerClassName,
}: HeadlessSelectProps) {
  return (
    <Select.Root value={value} onValueChange={onValueChange} disabled={disabled}>
      <Select.Trigger
        aria-label={ariaLabel}
        className={`inline-flex w-full items-center justify-between gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-200 outline-none transition hover:bg-zinc-800 focus:ring-1 focus:ring-zinc-500 disabled:cursor-not-allowed disabled:opacity-50 ${triggerClassName || ''}`}
      >
        <Select.Value placeholder={placeholder} />
        <Select.Icon>
          <ChevronDown className="h-4 w-4 text-zinc-500" />
        </Select.Icon>
      </Select.Trigger>
      <Select.Portal>
        <Select.Content
          position="popper"
          sideOffset={6}
          className="z-50 min-w-[220px] overflow-hidden rounded-md border border-zinc-700 bg-zinc-900 p-1 text-sm shadow-xl"
        >
          <Select.Viewport>
            {options.map((option) => (
              <Select.Item
                key={option.value}
                value={option.value}
                className="relative flex cursor-default select-none items-center rounded px-8 py-2 text-zinc-200 outline-none data-[highlighted]:bg-zinc-800"
              >
                <Select.ItemText>
                  {option.description ? (
                    <span className="flex flex-col">
                      <span>{option.label}</span>
                      <span className="text-xs text-zinc-500">{option.description}</span>
                    </span>
                  ) : (
                    option.label
                  )}
                </Select.ItemText>
                <Select.ItemIndicator className="absolute left-2 inline-flex items-center">
                  <Check className="h-4 w-4 text-zinc-400" />
                </Select.ItemIndicator>
              </Select.Item>
            ))}
          </Select.Viewport>
        </Select.Content>
      </Select.Portal>
    </Select.Root>
  );
}
