config {
    call_module_type = "all"
    force = false
    disabled_by_default = false
    ignore_module = {}
}

#Plugins
plugin "aws" {
    enabled = true
    version = "0.38.0"
    source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

#Rules
rule "terraform_comment_syntax" { enabled = true }
rule "terraform_deprecated_index" { enabled = true }
rule "terraform_documented_outputs" { enabled = true }
rule "terraform_naming_convention" { enabled = true }
rule "terraform_typed_variables" { enabled = true }
rule "terraform_unused_declarations" { enabled = true }
rule "terraform_unused_required_providers" { enabled = true }
rule "terraform_required_version" { enabled = false }
rule "terraform_required_providers" { enabled = false }
rule "terraform_module_version" {
  enabled = true
  exact = false # default
}
