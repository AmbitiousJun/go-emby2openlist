import type React from "react";
import { useState } from "react";

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
    <div className="collapse collapse-arrow bg-base-100 border border-base-300">
      <input
        type="checkbox"
        className="peer"
        checked={checked}
        onChange={(e) => setChecked(e.target.checked)}
      />
      <div className="collapse-title font-semibold after:start-5 after:end-auto pe-4 ps-12 bg-accent text-accent-content">
        {title}
      </div>
      <div className="collapse-content text-sm bg-base-200 text-base-content">
        {children}
      </div>
    </div>
  );
}
