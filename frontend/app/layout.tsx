import type { Metadata } from "next";
import { Geist, Geist_Mono, Instrument_Serif } from "next/font/google";
import "./globals.css";

const geistSans = Geist({ variable: "--font-geist-sans", subsets: ["latin"] });
const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});
const display = Instrument_Serif({
  variable: "--font-display",
  subsets: ["latin"],
  weight: "400",
  style: ["normal", "italic"],
});

export const metadata: Metadata = {
  title: "Aperture — one studio, every model",
  description:
    "Generate video, image and voice across every major model from one workspace. Bring your own keys or use ours.",
  // app/icon.svg and app/apple-icon.png are picked up by convention; this
  // names the one file that lives in public/ instead.
  icons: {
    icon: [{ url: "/icon.svg", type: "image/svg+xml" }],
    apple: [{ url: "/apple-icon.png", sizes: "180x180" }],
  },
  openGraph: {
    title: "Aperture — one studio, every model",
    description:
      "Video, stills and voice from a single prompt bar. Bring your own keys or use ours.",
    images: ["/logo.svg"],
  },
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    /* `js` is rendered here rather than added by a script at runtime. Adding it
       before hydration mutates an attribute React already rendered, which is a
       hydration mismatch; rendering it means server and client agree, and the
       hidden-until-animated state is still correct on the very first paint.
       The two ways that state could get stuck are handled below. */
    <html lang="en" className="js">
      <head>
        {/* Scripting disabled: nothing will ever animate, so nothing may hide. */}
        <noscript>
          <style>{`.reveal-target,.loop-panel{opacity:1!important;transform:none!important}`}</style>
        </noscript>
        {/* Scripting enabled but the app never booted (chunk failed, offline
            mid-load). This inline script always runs even when the main bundle
            does not, so it is the only thing that can recover that case. */}
        <script
          dangerouslySetInnerHTML={{
            __html:
              `window.setTimeout(function(){` +
              `if(!window.__apertureReady){document.documentElement.classList.remove("js")}` +
              `},5000)`,
          }}
        />
      </head>
      <body
        className={`${geistSans.variable} ${geistMono.variable} ${display.variable} grain antialiased`}
      >
        {children}
      </body>
    </html>
  );
}
