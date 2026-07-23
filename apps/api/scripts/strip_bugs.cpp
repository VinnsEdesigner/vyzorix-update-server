// strip_bugs.cpp
// Recursively walks a directory tree, strips // and /* */ comments that contain
// any BUG-01 … BUG-99 marker OR any case-variant of "fix" (fix/FIX/Fix/…).
// Writes the sanitised files back.

#include <iostream>
#include <fstream>
#include <sstream>
#include <string>
#include <vector>
#include <filesystem>
#include <regex>
#include <set>

namespace fs = std::filesystem;

// ---------------------------------------------------------------
// Returns true when 's' contains a target marker:
//   - BUG-NN  (e.g. BUG-01, BUG-99)
//   - fix/FIX/Fix (case-insensitive word-boundary match)
static bool contains_target_marker(const std::string& s) {
    static const std::regex bug_pat(R"(\bBUG-\d{2}\b)");
    if (std::regex_search(s, bug_pat)) return true;

    // Case-insensitive \bfix\b — matches "fix", "FIX", "Fix", "FiX", etc.
    static const std::regex fix_pat(R"(\bfix\b)", std::regex_constants::icase);
    if (std::regex_search(s, fix_pat)) return true;

    return false;
}

// ---------------------------------------------------------------
// Strip all /* … */ blocks whose text contains a BUG or fix marker.
// Uses regex_iterator to allow a callback-like lambda.
static std::string strip_ml_comments(const std::string& text) {
    static const std::regex ml_pat(R"(\/\*[\s\S]*?\*\/)");
    std::string result;
    std::sregex_iterator it(text.begin(), text.end(), ml_pat);
    std::sregex_iterator end;
    size_t last_pos = 0;
    for (; it != end; ++it) {
        // copy everything before this match unchanged
        result.append(text, last_pos, it->position() - last_pos);
        if (!contains_target_marker(it->str())) {
            result.append(it->str());   // keep the block
        }
        last_pos = it->position() + it->length();
    }
    result.append(text, last_pos, std::string::npos);
    return result;
}

// ---------------------------------------------------------------
// Remove trailing // comments that contain BUG or fix markers.
// Keeps the leading whitespace / indentation so the rest of the
// line stays valid code.
static std::string strip_sl_comments(const std::string& text) {
    std::istringstream in(text);
    std::ostringstream out;
    std::string line;
    while (std::getline(in, line)) {
        size_t pos = line.find("//");
        if (pos != std::string::npos) {
            std::string comment = line.substr(pos);
            if (contains_target_marker(comment)) {
                out << line.substr(0, pos) << "\n";   // strip the comment
                continue;
            }
        }
        out << line << "\n";
    }
    return out.str();
}

// ---------------------------------------------------------------
// Process a single file: strip bug/fix comments and write back.
// Returns true only if at least one matching comment was removed.
static bool process_file(const fs::path& p) {
    std::ifstream in(p, std::ios::binary);
    if (!in) {
        std::cerr << "[SKIP] cannot read: " << p << "\n";
        return false;
    }

    std::string content((std::istreambuf_iterator<char>(in)),
                         std::istreambuf_iterator<char>());
    in.close();

    // Track whether the original file had any target marker BEFORE we strip.
    bool had_marker = contains_target_marker(content);

    // 1. Strip multi-line /* … */ comments containing markers.
    content = strip_ml_comments(content);

    // 2. Strip single-line // comments containing markers.
    content = strip_sl_comments(content);

    // Only write if we actually removed something.
    if (!had_marker) return false;

    std::ofstream out(p, std::ios::binary | std::ios::trunc);
    if (!out) {
        std::cerr << "[ERROR] cannot write: " << p << "\n";
        return false;
    }
    out << content;
    out.close();

    std::cout << "[STRIPPED] " << p << "\n";
    return true;
}

// ---------------------------------------------------------------
// Files / directories to ignore during the walk.
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

    if (p.filename() == "strip_bugs.cpp") return true;
    if (s_ignore_dirs.count(p.filename().string())) return true;
    if (s_ignore_ext.count(p.extension().string())) return true;

    std::string name = p.filename().string();
    if (!name.empty() && name.front() == '.') return true;

    return false;
}

// ---------------------------------------------------------------
int main(int argc, char* argv[]) {
    if (argc < 2) {
        std::cerr << "Usage: " << argv[0] << " <root_dir> [--dry-run]\n";
        return 1;
    }

    fs::path root = argv[1];
    bool dry_run = (argc >= 3 && std::string(argv[2]) == "--dry-run");

    if (!fs::exists(root) || !fs::is_directory(root)) {
        std::cerr << "Error: not a valid directory: " << root << "\n";
        return 1;
    }

    std::vector<fs::path> changed;

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
            if (dry_run) {
                std::ifstream in(ep);
                std::string content((std::istreambuf_iterator<char>(in)),
                                     std::istreambuf_iterator<char>());
                if (contains_target_marker(content)) {
                    std::cout << "[DRY-RUN] " << ep << "\n";
                }
            } else {
                if (process_file(ep))
                    changed.push_back(ep);
            }
        }
    }

    if (dry_run) {
        std::cout << "\n(Dry run — no files were modified)\n";
    } else {
        std::cout << "\nDone. " << changed.size() << " file(s) modified.\n";
    }

    return 0;
}
