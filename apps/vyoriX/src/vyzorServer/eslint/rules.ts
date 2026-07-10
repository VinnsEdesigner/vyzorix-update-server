        const forbiddenImports = [
          { pattern: /^react$/, name: "React" },
          { pattern: /^react-dom/, name: "react-dom" },
          { pattern: /^@\/components/, name: "@/components" },
          { pattern: /^@\/hooks/, name: "@/hooks" },
          { pattern: /^@\/vyzorServer\/(?!api\/)/, name: "@/vyzorServer (non-API)" },
        ];