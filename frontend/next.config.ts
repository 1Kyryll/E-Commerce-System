import type { NextConfig } from "next";

const GATEWAY = process.env.GATEWAY_URL ?? "http://localhost:8080";

const nextConfig: NextConfig = {
  // Emit a self-contained server bundle (.next/standalone) for a slim
  // production container image.
  output: "standalone",
  async rewrites() {
    return [{ source: "/api/:path*", destination: `${GATEWAY}/:path*` }];
  },
};

export default nextConfig;
