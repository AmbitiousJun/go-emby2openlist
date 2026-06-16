import { ChevronDownIcon } from "lucide-react";
import type React from "react";
import { useState } from "react";
import { Button } from "~/components/ui/button";
import { Card, CardContent } from "~/components/ui/card";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "~/components/ui/collapsible";

export type CommonCollapseProps = {
  title: string;
  defaultChecked?: boolean;
  children?: React.ReactNode;
};

export default function CommonCollapse({
  title,
  defaultChecked,
  children,
}: CommonCollapseProps) {
  const [checked, setChecked] = useState(defaultChecked ?? false);

  return (
    <Card size="sm" className="mx-auto w-full">
      <CardContent>
        <Collapsible
          className="rounded-md data-[state=open]:bg-muted data-[state=open]:pt-2.5"
          open={checked}
          onOpenChange={setChecked}
        >
          <CollapsibleTrigger asChild>
            <Button variant="ghost" className="group w-full text-lg px-2.5">
              {title}
              <ChevronDownIcon className="ml-auto group-data-[state=open]:rotate-180" />
            </Button>
          </CollapsibleTrigger>
          <CollapsibleContent className="w-full flex flex-col items-start gap-2 p-5 pt-0 text-sm">
            {children}
          </CollapsibleContent>
        </Collapsible>
      </CardContent>
    </Card>
  );
}
