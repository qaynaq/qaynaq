import React, { useCallback, useEffect } from "react";
import { Dialog, DialogTitle, DialogContent } from "@/components/ui/dialog";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { useNavigate } from "react-router-dom";
import {
  Moon,
  Sun,
  LayoutDashboard,
  Component,
  Cpu,
  MemoryStick,
  Layers,
  ScanLine,
  Waypoints,
  Monitor,
} from "lucide-react";
import { useTheme } from "next-themes";

interface CommandPaletteProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CommandPalette({ open, onOpenChange }: CommandPaletteProps) {
  const navigate = useNavigate();
  const { setTheme } = useTheme();
  const [inputValue, setInputValue] = React.useState("");

  const runCommand = useCallback((command: () => void) => {
    onOpenChange(false);
    command();
  }, [onOpenChange]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTitle className="text-lg font-semibold hidden">
        Command Palette
      </DialogTitle>
      <DialogContent className="p-0 gap-0 max-w-[640px] overflow-hidden rounded-lg">
        <Command className="rounded-lg border shadow-md">
          <CommandInput
            placeholder="Search..."
            value={inputValue}
            onValueChange={setInputValue}
          />
          <CommandList>
            <CommandEmpty>No results found.</CommandEmpty>
            <CommandGroup heading="General">
              <CommandItem
                onSelect={() => runCommand(() => navigate("/"))}
              >
                <LayoutDashboard className="mr-2 h-4 w-4" />
                Dashboard
              </CommandItem>
              <CommandItem
                onSelect={() => runCommand(() => navigate("/flows/new"))}
              >
                <Waypoints className="mr-2 h-4 w-4" />
                Flows
              </CommandItem>
              <CommandItem
                onSelect={() => runCommand(() => navigate("/workers"))}
              >
                <Cpu className="mr-2 h-4 w-4" />
                Workers
              </CommandItem>
            </CommandGroup>
            <CommandGroup heading="Other">
              <CommandItem
                onSelect={() => runCommand(() => navigate("/caches"))}
              >
                <MemoryStick className="mr-2 h-4 w-4" />
                Caches
              </CommandItem>
              <CommandItem
                onSelect={() => runCommand(() => navigate("/buffers"))}
              >
                <Layers className="mr-2 h-4 w-4" />
                Buffers
              </CommandItem>
              <CommandItem
                onSelect={() => runCommand(() => navigate("/scanners"))}
              >
                <ScanLine className="mr-2 h-4 w-4" />
                Scanners
              </CommandItem>
            </CommandGroup>
            <CommandGroup heading="Theme">
              <CommandItem
                onSelect={() => runCommand(() => setTheme("light"))}
              >
                <Sun className="mr-2 h-4 w-4" />
                Light
              </CommandItem>
              <CommandItem
                onSelect={() => runCommand(() => setTheme("dark"))}
              >
                <Moon className="mr-2 h-4 w-4" />
                Dark
              </CommandItem>
              <CommandItem
                onSelect={() => runCommand(() => setTheme("system"))}
              >
                <Monitor className="mr-2 h-4 w-4" />
                System
              </CommandItem>
            </CommandGroup>
          </CommandList>
        </Command>
      </DialogContent>
    </Dialog>
  );
}

