import type { Metadata } from "next";
import { Geist, Geist_Mono, Instrument_Serif } from "next/font/google";
import "./globals.css";

const geistSans = Geist({ variable: "--font-geist-sans", subsets: ["latin"] });
const geistMono = Geist_Mono({ variable: "--font-geist-mono", subsets: ["latin"] });
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
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" className="js">
      <body
        className={`${geistSans.variable} ${geistMono.variable} ${display.variable} grain antialiased`}
      >
        {children}
      </body>
    </html>
  );
}
