<?php

class __Mustache_7779e512fd2c33f860b4bd4d826ec9c8 extends Mustache_Template
{
    private $lambdaHelper;

    public function renderInternal(Mustache_Context $context, $indent = '')
    {
        $this->lambdaHelper = new Mustache_LambdaHelper($this->mustache, $context);
        $buffer = '';

        $buffer .= $indent . '<div class="centered-menu">
';
        $buffer .= $indent . '    <div class="card">
';
        $buffer .= $indent . '        <form id="cartridge-registration-form" class="form-horizontal">
';
        $buffer .= $indent . '            <p class="lead text-center">';
        $value = $context->find('str');
        $buffer .= $this->sectionBcc9f3a2d2213bff929a1124fd2d6dc7($context, $indent, $value);
        $buffer .= '</p>
';
        $buffer .= $indent . '            <p class="text-center">';
        $value = $context->find('str');
        $buffer .= $this->section27c9075a744fb69eac54959e08e8fe17($context, $indent, $value);
        $buffer .= '</p>
';
        $buffer .= $indent . '            <div class="control-group">
';
        $buffer .= $indent . '                <div class="control-label">
';
        $buffer .= $indent . '                    <label for="registration-key" style="display: inline-block">';
        $value = $context->find('str');
        $buffer .= $this->section5612caaf36ad1fbb4e8b71ea965c325d($context, $indent, $value);
        $buffer .= '</label>
';
        $value = $context->find('keyhelp');
        $buffer .= $this->sectionAe627446a49ca666bd6263dd5f3c4c07($context, $indent, $value);
        $buffer .= $indent . '                </div>
';
        $buffer .= $indent . '                <div class="controls">
';
        $buffer .= $indent . '                    <input name="tool-key"
';
        $buffer .= $indent . '                        class="input-block-level form-control"
';
        $buffer .= $indent . '                        type="text"
';
        $buffer .= $indent . '                        id="registration-key">
';
        $buffer .= $indent . '                </div>
';
        $buffer .= $indent . '            </div>
';
        $buffer .= $indent . '            <div class="control-group">
';
        $buffer .= $indent . '                <div class ="control-label">
';
        $buffer .= $indent . '                    <label for="registration-secret" style="display: inline-block">';
        $value = $context->find('str');
        $buffer .= $this->section3e6b7665cff7f26bf6bf125132a0e267($context, $indent, $value);
        $buffer .= '</label>
';
        $value = $context->find('secrethelp');
        $buffer .= $this->sectionAe627446a49ca666bd6263dd5f3c4c07($context, $indent, $value);
        $buffer .= $indent . '                </div>
';
        $buffer .= $indent . '                <div class="controls">
';
        $buffer .= $indent . '                    <input name="tool-secret"
';
        $buffer .= $indent . '                        class="input-block-level form-control"
';
        $buffer .= $indent . '                        type="text"
';
        $buffer .= $indent . '                        id="registration-secret">
';
        $buffer .= $indent . '                </div>
';
        $buffer .= $indent . '            </div>
';
        $buffer .= $indent . '            <div class="control-group mt-1">
';
        $buffer .= $indent . '                <div class="controls">
';
        $buffer .= $indent . '                    <button id="cartridge-registration-submit" type="submit" class="btn btn-success">
';
        $buffer .= $indent . '                        <span class="btn-text">';
        $value = $context->find('str');
        $buffer .= $this->section5200e3c03061df0fcc2d285702091399($context, $indent, $value);
        $buffer .= '</span>
';
        $buffer .= $indent . '                        <span class="btn-loader">
';
        if ($partial = $this->mustache->loadPartial('mod_lti/loader')) {
            $buffer .= $partial->renderInternal($context, $indent . '                            ');
        }
        $buffer .= $indent . '                        </span>
';
        $buffer .= $indent . '                    </button>
';
        $buffer .= $indent . '                    <button id="cartridge-registration-cancel" type="button" class="btn">
';
        $buffer .= $indent . '                        <span class="btn-text">';
        $value = $context->find('str');
        $buffer .= $this->section48889b9f3f273ba8c7c463afc8a04b66($context, $indent, $value);
        $buffer .= '</span>
';
        $buffer .= $indent . '                        <span class="btn-loader">
';
        if ($partial = $this->mustache->loadPartial('mod_lti/loader')) {
            $buffer .= $partial->renderInternal($context, $indent . '                            ');
        }
        $buffer .= $indent . '                        </span>
';
        $buffer .= $indent . '                    </button>
';
        $buffer .= $indent . '                </div>
';
        $buffer .= $indent . '            </div>
';
        $buffer .= $indent . '        </form>
';
        $buffer .= $indent . '    </div>
';
        $buffer .= $indent . '</div>
';
        $value = $context->find('js');
        $buffer .= $this->sectionB6e1a9dfbef284e4ad52cb09e96a17d7($context, $indent, $value);

        return $buffer;
    }

    private function sectionBcc9f3a2d2213bff929a1124fd2d6dc7(Mustache_Context $context, $indent, $value)
    {
        $buffer = '';
    
        if (!is_string($value) && is_callable($value)) {
            $source = ' enterkeyandsecret, mod_lti ';
            $result = (string) call_user_func($value, $source, $this->lambdaHelper);
            $buffer .= $result;
        } elseif (!empty($value)) {
            $values = $this->isIterable($value) ? $value : array($value);
            foreach ($values as $value) {
                $context->push($value);
                
                $buffer .= ' enterkeyandsecret, mod_lti ';
                $context->pop();
            }
        }
    
        return $buffer;
    }

    private function section27c9075a744fb69eac54959e08e8fe17(Mustache_Context $context, $indent, $value)
    {
        $buffer = '';
    
        if (!is_string($value) && is_callable($value)) {
            $source = ' enterkeyandsecret_help, mod_lti ';
            $result = (string) call_user_func($value, $source, $this->lambdaHelper);
            $buffer .= $result;
        } elseif (!empty($value)) {
            $values = $this->isIterable($value) ? $value : array($value);
            foreach ($values as $value) {
                $context->push($value);
                
                $buffer .= ' enterkeyandsecret_help, mod_lti ';
                $context->pop();
            }
        }
    
        return $buffer;
    }

    private function section5612caaf36ad1fbb4e8b71ea965c325d(Mustache_Context $context, $indent, $value)
    {
        $buffer = '';
    
        if (!is_string($value) && is_callable($value)) {
            $source = ' resourcekey_admin, mod_lti ';
            $result = (string) call_user_func($value, $source, $this->lambdaHelper);
            $buffer .= $result;
        } elseif (!empty($value)) {
            $values = $this->isIterable($value) ? $value : array($value);
            foreach ($values as $value) {
                $context->push($value);
                
                $buffer .= ' resourcekey_admin, mod_lti ';
                $context->pop();
            }
        }
    
        return $buffer;
    }

    private function sectionAe627446a49ca666bd6263dd5f3c4c07(Mustache_Context $context, $indent, $value)
    {
        $buffer = '';
    
        if (!is_string($value) && is_callable($value)) {
            $source = '
                        {{>core/help_icon}}
                    ';
            $result = (string) call_user_func($value, $source, $this->lambdaHelper);
            $buffer .= $result;
        } elseif (!empty($value)) {
            $values = $this->isIterable($value) ? $value : array($value);
            foreach ($values as $value) {
                $context->push($value);
                
                if ($partial = $this->mustache->loadPartial('core/help_icon')) {
                    $buffer .= $partial->renderInternal($context, $indent . '                        ');
                }
                $context->pop();
            }
        }
    
        return $buffer;
    }

    private function section3e6b7665cff7f26bf6bf125132a0e267(Mustache_Context $context, $indent, $value)
    {
        $buffer = '';
    
        if (!is_string($value) && is_callable($value)) {
            $source = ' password_admin, mod_lti ';
            $result = (string) call_user_func($value, $source, $this->lambdaHelper);
            $buffer .= $result;
        } elseif (!empty($value)) {
            $values = $this->isIterable($value) ? $value : array($value);
            foreach ($values as $value) {
                $context->push($value);
                
                $buffer .= ' password_admin, mod_lti ';
                $context->pop();
            }
        }
    
        return $buffer;
    }

    private function section5200e3c03061df0fcc2d285702091399(Mustache_Context $context, $indent, $value)
    {
        $buffer = '';
    
        if (!is_string($value) && is_callable($value)) {
            $source = ' savechanges ';
            $result = (string) call_user_func($value, $source, $this->lambdaHelper);
            $buffer .= $result;
        } elseif (!empty($value)) {
            $values = $this->isIterable($value) ? $value : array($value);
            foreach ($values as $value) {
                $context->push($value);
                
                $buffer .= ' savechanges ';
                $context->pop();
            }
        }
    
        return $buffer;
    }

    private function section48889b9f3f273ba8c7c463afc8a04b66(Mustache_Context $context, $indent, $value)
    {
        $buffer = '';
    
        if (!is_string($value) && is_callable($value)) {
            $source = ' cancel ';
            $result = (string) call_user_func($value, $source, $this->lambdaHelper);
            $buffer .= $result;
        } elseif (!empty($value)) {
            $values = $this->isIterable($value) ? $value : array($value);
            foreach ($values as $value) {
                $context->push($value);
                
                $buffer .= ' cancel ';
                $context->pop();
            }
        }
    
        return $buffer;
    }

    private function sectionB6e1a9dfbef284e4ad52cb09e96a17d7(Mustache_Context $context, $indent, $value)
    {
        $buffer = '';
    
        if (!is_string($value) && is_callable($value)) {
            $source = '
    require([\'mod_lti/cartridge_registration_form\'], function(registration) {
        registration.init();
    });
';
            $result = (string) call_user_func($value, $source, $this->lambdaHelper);
            $buffer .= $result;
        } elseif (!empty($value)) {
            $values = $this->isIterable($value) ? $value : array($value);
            foreach ($values as $value) {
                $context->push($value);
                
                $buffer .= $indent . '    require([\'mod_lti/cartridge_registration_form\'], function(registration) {
';
                $buffer .= $indent . '        registration.init();
';
                $buffer .= $indent . '    });
';
                $context->pop();
            }
        }
    
        return $buffer;
    }

}
