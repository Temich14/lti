<?php

class __Mustache_ad8f026f92992cc0523a63454db96207 extends Mustache_Template
{
    private $lambdaHelper;

    public function renderInternal(Mustache_Context $context, $indent = '')
    {
        $this->lambdaHelper = new Mustache_LambdaHelper($this->mustache, $context);
        $buffer = '';

        $buffer .= $indent . '<div id="tool-card-container">
';
        $buffer .= $indent . '    ';
        $value = $context->find('proxies');
        if (empty($value)) {
            
            $value = $context->find('tools');
            if (empty($value)) {
                
                $buffer .= '
';
                $buffer .= $indent . '        <p>';
                $value = $context->find('str');
                $buffer .= $this->section480b8c0397cb86d33ef2566f1fb0f370($context, $indent, $value);
                $buffer .= '</p>
';
                $buffer .= $indent . '    ';
            }
        }
        $buffer .= '
';
        $value = $context->find('proxies');
        $buffer .= $this->sectionCf472df0108766e62249a6471c80e50b($context, $indent, $value);
        $value = $context->find('tools');
        $buffer .= $this->section9341771ec1e55aa1f87479c2f2632b00($context, $indent, $value);
        $buffer .= $indent . '</div>
';

        return $buffer;
    }

    private function section480b8c0397cb86d33ef2566f1fb0f370(Mustache_Context $context, $indent, $value)
    {
        $buffer = '';
    
        if (!is_string($value) && is_callable($value)) {
            $source = ' no_lti_tools, mod_lti ';
            $result = (string) call_user_func($value, $source, $this->lambdaHelper);
            $buffer .= $result;
        } elseif (!empty($value)) {
            $values = $this->isIterable($value) ? $value : array($value);
            foreach ($values as $value) {
                $context->push($value);
                
                $buffer .= ' no_lti_tools, mod_lti ';
                $context->pop();
            }
        }
    
        return $buffer;
    }

    private function sectionCf472df0108766e62249a6471c80e50b(Mustache_Context $context, $indent, $value)
    {
        $buffer = '';
    
        if (!is_string($value) && is_callable($value)) {
            $source = '
        {{> mod_lti/tool_proxy_card }}
    ';
            $result = (string) call_user_func($value, $source, $this->lambdaHelper);
            $buffer .= $result;
        } elseif (!empty($value)) {
            $values = $this->isIterable($value) ? $value : array($value);
            foreach ($values as $value) {
                $context->push($value);
                
                if ($partial = $this->mustache->loadPartial('mod_lti/tool_proxy_card')) {
                    $buffer .= $partial->renderInternal($context, $indent . '        ');
                }
                $context->pop();
            }
        }
    
        return $buffer;
    }

    private function section9341771ec1e55aa1f87479c2f2632b00(Mustache_Context $context, $indent, $value)
    {
        $buffer = '';
    
        if (!is_string($value) && is_callable($value)) {
            $source = '
        {{> mod_lti/tool_card }}
    ';
            $result = (string) call_user_func($value, $source, $this->lambdaHelper);
            $buffer .= $result;
        } elseif (!empty($value)) {
            $values = $this->isIterable($value) ? $value : array($value);
            foreach ($values as $value) {
                $context->push($value);
                
                if ($partial = $this->mustache->loadPartial('mod_lti/tool_card')) {
                    $buffer .= $partial->renderInternal($context, $indent . '        ');
                }
                $context->pop();
            }
        }
    
        return $buffer;
    }

}
