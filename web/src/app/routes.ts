import {
  type RouteConfig,
  index,
  layout,
  route,
} from "@react-router/dev/routes";

export default [
  layout("routes/layout.tsx", [
    index("routes/index.tsx"),
    route(
      "api/openlist_local_tree",
      "routes/api/openlist_local_tree/index.tsx",
    ),
  ]),
] satisfies RouteConfig;
