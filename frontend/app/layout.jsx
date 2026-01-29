import "./globals.css";

export const metadata = {
  title: "RAG Document Search",
  description: "Multi-user RAG document search",
};

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <body>
        <div className="container">{children}</div>
      </body>
    </html>
  );
}
