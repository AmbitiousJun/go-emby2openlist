import { type RouteConfig, index, layout, route } from "@react-router/dev/routes";

export default [
    layout("routes/layout.tsx", [
        index("routes/index.tsx"),
        route("sub1", "routes/sub1.tsx"),
        route("sub2", "routes/sub2.tsx"),
    ])
] satisfies RouteConfig;
