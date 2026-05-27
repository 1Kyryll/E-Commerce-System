import type { NextConfig } from "next";

const GATEWAY = process.env.GATEWAY_URL ?? "http://localhost:8080";

const nextConfig: NextConfig = {
  async rewrites() {
    return [{ source: "/api/:path*", destination: `${GATEWAY}/:path*` }];
  },
};

export default nextConfig;
