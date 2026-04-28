<?php

class __Mustache_5fba4fb6928fc054876475579f6de5e7 extends Mustache_Template
{
    private $lambdaHelper;

    public function renderInternal(Mustache_Context $context, $indent = '')
    {
        $this->lambdaHelper = new Mustache_LambdaHelper($this->mustache, $context);
        $buffer = '';

        $buffer .= $indent . '<span class="loader">
';
        $buffer .= $indent . '    ';
        $value = $context->find('pix');
        $buffer .= $this->sectionC8244e3ff2e632f5dcf1816fe4378766($context, $indent, $value);
        $buffer .= '
';
        $buffer .= $indent . '</span>
';

        return $buffer;
    }

    private function sectionD3502cb8652b5ca94b45d7b52a2201ac(Mustache_Context $context, $indent, $value)
    {
        $buffer = '';
    
        if (!is_string($value) && is_callable($value)) {
            $source = ' loadinghelp, moodle ';
            $result = (string) call_user_func($value, $source, $this->lambdaHelper);
            $buffer .= $result;
        } elseif (!empty($value)) {
            $values = $this->isIterable($value) ? $value : array($value);
            foreach ($values as $value) {
                $context->push($value);
                
                $buffer .= ' loadinghelp, moodle ';
                $context->pop();
            }
        }
    
        return $buffer;
    }

    private function sectionC8244e3ff2e632f5dcf1816fe4378766(Mustache_Context $context, $indent, $value)
    {
        $buffer = '';
    
        if (!is_string($value) && is_callable($value)) {
            $source = ' i/loading, core, {{#str}} loadinghelp, moodle {{/str}} ';
            $result = (string) call_user_func($value, $source, $this->lambdaHelper);
            $buffer .= $result;
        } elseif (!empty($value)) {
            $values = $this->isIterable($value) ? $value : array($value);
            foreach ($values as $value) {
                $context->push($value);
                
                $buffer .= ' i/loading, core, ';
                $value = $context->find('str');
                $buffer .= $this->sectionD3502cb8652b5ca94b45d7b52a2201ac($context, $indent, $value);
                $buffer .= ' ';
                $context->pop();
            }
        }
    
        return $buffer;
    }

}
