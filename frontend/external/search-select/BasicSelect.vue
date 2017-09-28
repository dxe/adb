<template>
  <!-- SAMER: do we need isError? -->
  <div class="chosen-container chosen-container-single"
       :class="containerClass"
       :style="inheritStyle"
       @mousedown="mousedown"
       @click="toggleOptions">
    <div class="chosen-single chosen-single-with-deselect"
         :class="textClass">
      <span>{{inputText}}</span>
      <div><b></b></div>
    </div>
    <div class="chosen-drop">
      <div class="chosen-search">
        <input class="chosen-search-input"
               autocomplete="off"
               tabindex="0"
               v-model="searchText"
               ref="input"
               type="text"
               @click.stop
               @blur="blurInput"
               @keydown.up="prevItem"
               @keydown.down="nextItem"
               @keydown.enter.prevent=""
               @keyup.enter.prevent="enterItem"
               @keydown.delete="deleteTextOrItem"
               />
      </div>

      <ul class="chosen-results"
           ref="menu"
           @mousedown.prevent
           :class="menuClass"
           :style="menuStyle"
           tabindex="-1">
        <template v-for="(option, idx) in filteredOptions">
          <li class="active-result"
               :class="{ 'highlighted': option.selected || pointer === idx }"
               @click.stop="selectItem(option)"
               @mousedown="mousedown"
               @mouseenter="pointerSet(idx)">
            {{option.text}}
          </li>
        </template>
      </ul>
    </div>
  </div>
</template>

<script>
  /* event : select */
  import common from './common'
  import commonMixin from './commonMixin'

  export default {
    mixins: [commonMixin],
    props: {
      options: {
        type: Array
      },
      selectedOption: {
        type: Object,
        default: () => { return { value: '', text: '' } }
      },
      extraData: {
        type: Object,
        default: () => { return {} },
      },
      inheritStyle: {
        type: String,
        default: () => { return ''; },
      },
    },
    data () {
      return {
        showMenu: false,
        searchText: '',
        mousedownState: false, // mousedown on option menu
        pointer: 0,
      }
    },
    watch: {
      filteredOptions () {
        this.pointerAdjust()
      }
    },
    computed: {
      inputText () {
        let text = this.placeholder
        if (this.selectedOption.text) {
          text = this.selectedOption.text
        }
        return text
      },
      textClass () {
        if (!this.selectedOption.text && this.placeholder) {
          return 'chosen-default'
        } else {
          return ''
        }
      },
      containerClass() {
        return {
          'chosen-with-drop': this.showMenu,
          'chosen-container-active': this.showMenu,
        };
      },
      menuClass () {
        return {
          visible: this.showMenu,
          hidden: !this.showMenu
        }
      },
      menuStyle () {
        return {
          display: this.showMenu ? 'block' : 'none'
        }
      },
      filteredOptions () {
        if (this.searchText) {
          return this.options.filter((option) => {
            try {
              return this.filterPredicate(option.text, this.searchText)
            } catch (e) {
              return true
            }
          })
        } else {
          return this.options
        }
      }
    },
    methods: {
      deleteTextOrItem () {
        if (!this.searchText && this.selectedOption) {
          this.selectItem({})
          this.openOptions()
        }
      },
      toggleOptions() {
        common.toggleOptions(this);
      },
      openOptions () {
        common.openOptions(this)
      },
      blurInput () {
        common.blurInput(this)
      },
      closeOptions () {
        common.closeOptions(this)
      },
      prevItem () {
        common.prevItem(this)
      },
      nextItem () {
        common.nextItem(this)
      },
      enterItem () {
        common.enterItem(this)
      },
      pointerSet (index) {
        common.pointerSet(this, index)
      },
      pointerAdjust () {
        common.pointerAdjust(this)
      },
      mousedown () {
        common.mousedown(this)
      },
      selectItem (option) {
        this.searchText = '' // reset text when select item
        this.closeOptions()
        this.$emit('select', option, this.extraData)
      }
    }
  }
</script>

<style src='bootstrap-chosen/bootstrap-chosen.css'></style>
<style>
  /* Menu Item Hover */
  .ui.dropdown .menu > .item:hover {
    background: none transparent !important;
  }
  
  /* Menu Item Hover for Key event */
  .ui.dropdown .menu > .item.current {
    background: rgba(0, 0, 0, 0.05) !important;
  }
</style>
