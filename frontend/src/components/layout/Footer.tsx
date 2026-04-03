import { BookOpen } from "lucide-react";

export function Footer() {
  return (
    <footer className="border-t py-6 mt-auto">
      <div className="container flex flex-col sm:flex-row items-center justify-between gap-4">
        <div className="flex items-center gap-2 text-muted-foreground">
          <BookOpen className="h-4 w-4" />
          <span className="text-sm">Clean Architecture Monolith by L3XT</span>
        </div>
        <a className="text-sm text-muted-foreground" href="https://github.com/l3xt/">Мой GitHub</a>
      </div>
    </footer>
  );
}
