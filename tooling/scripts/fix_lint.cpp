// fix_lint.cpp
// Batch fixes common linter issues:
// 1. wastedassign: Remove unused variable initializations that are immediately reassigned

#include <iostream>
#include <fstream>
#include <sstream>
#include <string>
#include <vector>
#include <filesystem>
#include <regex>
#include <set>

namespace fs = std::filesystem;

static std::vector<std::string> g_fixed_files;
static int g_wastedassign_fixes = 0;

// ---------------------------------------------------------------
// Fix wastedassign: Remove unused variable initialization
// Pattern: varName := value followed by varName = differentValue (same block)
// ONLY fixes simple numeric/boolean initializations that are immediately overwritten
static std::string fix_wastedassign(const std::string& text) {
    std::vector<std::string> lines;
    std::istringstream in(text);
    std::string line;
    while (std::getline(in, line)) {
        lines.push_back(line);
    }

    bool modified = false;
    for (size_t i = 0; i + 1 < lines.size(); i++) {
        // Only match simple patterns like: remaining := 0
        // NOT: role := operator.RoleOperator (this is a real value)
        std::smatch match;
        if (std::regex_match(lines[i], match,
            std::regex(R"(^\s*(\w+)\s*:=\s*(0|1|false|true)\s*$)", 
                       std::regex_constants::ECMAScript))) {
            std::string var_name = match[1].str();
            
            // Only fix known patterns: remaining, count, etc. (not role)
            if (var_name == "remaining" || var_name == "count") {
                // Search within 30 lines for reassignment
                for (size_t j = i + 1; j < std::min(i + 30, lines.size()); j++) {
                    std::string search_for = var_name + " =";
                    if (lines[j].find(search_for) != std::string::npos) {
                        lines[i].clear();  // Empty the line
                        modified = true;
                        g_wastedassign_fixes++;
                        std::cout << "  Fixed wastedassign: " << var_name 
                                  << " at line " << (i+1) << "\n";
                        break;
                    }
                }
            }
        }
    }

    if (!modified) return text;

    std::ostringstream out;
    for (const auto& l : lines) {
        out << l << "\n";
    }
    return out.str();
}

// ---------------------------------------------------------------
// Main processing function
static bool process_file(const fs::path& p) {
    std::ifstream in(p, std::ios::binary);
    if (!in) return false;

    std::string content((std::istreambuf_iterator<char>(in)),
                         std::istreambuf_iterator<char>());
    in.close();

    std::string original = content;

    // Apply fixes
    content = fix_wastedassign(content);

    if (content == original) return false;

    std::ofstream out(p, std::ios::binary | std::ios::trunc);
    if (!out) return false;
    out << content;
    out.close();

    g_fixed_files.push_back(p.string());
    return true;
}

// ---------------------------------------------------------------
static const std::set<std::string> s_ignore_dirs = {
    ".git", ".svn", "node_modules", "vendor", "__pycache__",
    "dist", "build", "target", ".venv", "venv",
};

static const std::set<std::string> s_ignore_ext = {
    ".zip", ".tar", ".gz", ".bz2", ".xz",
    ".png", ".jpg", ".jpeg", ".gif", ".ico", ".woff", ".woff2",
    ".pdf", ".doc", ".docx",
};

static bool should_ignore(const fs::path& p) {
    if (p.filename().empty()) return false;
    if (s_ignore_dirs.count(p.filename().string())) return true;
    if (s_ignore_ext.count(p.extension().string())) return true;
    std::string name = p.filename().string();
    if (!name.empty() && name.front() == '.') return true;
    return false;
}

// ---------------------------------------------------------------
int main(int argc, char* argv[]) {
    if (argc < 2) {
        std::cerr << "Usage: " << argv[0] << " <root_dir>\n";
        return 1;
    }

    fs::path root = argv[1];
    if (!fs::exists(root) || !fs::is_directory(root)) {
        std::cerr << "Error: not a valid directory: " << root << "\n";
        return 1;
    }

    for (auto it = fs::recursive_directory_iterator(root);
         it != fs::recursive_directory_iterator(); ++it) {
        const fs::path& ep = it->path();

        if (it->is_directory()) {
            if (should_ignore(ep))
                it.disable_recursion_pending();
            continue;
        }

        if (it->is_regular_file()) {
            if (should_ignore(ep)) continue;
            // Only process .go files
            if (ep.extension() != ".go") continue;
            process_file(ep);
        }
    }

    std::cout << "\nFixed " << g_fixed_files.size() << " file(s)\n";
    std::cout << "wastedassign fixes: " << g_wastedassign_fixes << "\n";

    return 0;
}
