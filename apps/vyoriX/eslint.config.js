          patterns: [
            { group: ["@/vyzorServer/api/**"], message: "[UIâData] Use hooks from @/hooks instead." },
            { group: ["@/domain/**"], message: "[UIâDomain] Use transformed data from hooks." },
            { group: ["@/components/**"], message: "[HooksâUI] Hooks must be pure logic." },
            { group: ["@/vyzorServer/**"], message: "[DomainâData] Domain must be pure types/transforms." },
            { group: ["@/hooks/**"], message: "[DomainâHooks] Domain must be pure types/transforms." },
            { group: ["@/components/**"], message: "[DomainâUI] Domain must be pure types/transforms." },
          ],