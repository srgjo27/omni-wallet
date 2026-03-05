import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Produce a minimal self-contained build (needed by Dockerfile runner stage)
  output: "standalone",

  // All backend traffic is routed through the API Gateway at port 8080.
  // The frontend never calls user-service or wallet-service directly.
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
